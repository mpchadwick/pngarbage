package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/aymerick/douceur/parser"
)

const version = "0.2.0"

var url string

var usageStr = `To scan a URL...

	Just supply the URL as follows

	./pngarbage -url="http://localhost:8080"

To check the version...

	Just run the version command

	./pngarbage version
`

func showUsage() {
	printBanner()
	fmt.Println(usageStr)
	os.Exit(1)
}

func showVersion() {
	fmt.Println(version)
	os.Exit(1)
}

func printBanner() {
	fmt.Println("===========================")
	fmt.Println("> pngarbage")
	fmt.Println("===========================")
}

/**
 * Find all the PNGs on a web page
 */
func findPngs(url string) []string {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	var pngs []string

	doc.Find("style").Each(func(i int, s *goquery.Selection) {
		html, err := s.Html()
		if err != nil {
			log.Fatal(err)
		}

		stylesheet, err := parser.Parse(html)
		if err != nil {
			log.Fatal(err)
		}

		for _, rule := range stylesheet.Rules {
			for _, declaration := range rule.Declarations {
				if declaration.Property == "background-image" {
					src := processBackgroundImageValue(declaration.Value)
					matched, _ := regexp.MatchString(".png$", src)
					if matched {
						pngs = append(pngs, src)
					}
				}
			}
		}
	})

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		matched, _ := regexp.MatchString(".png$", src)
		if matched {
			pngs = append(pngs, src)
		}
	})

	return pngs
}

/**
 *
 */
func processBackgroundImageValue(value string) string {
	// See: https://stackoverflow.com/a/20857703
	re := regexp.MustCompile(`(?:\(['"]?)(.*?)(?:['"]?\))`)

	matches := re.FindStringSubmatch(value)

	return matches[1]
}

/**
 * Prepare the request url from the PNG's src attribute
 */
func imgUrl(src string, url string) string {
	hasScheme, _ := regexp.MatchString("^https?://", src)
	if hasScheme {
		return src
	}

	srcHasInitialSlash, _ := regexp.MatchString("^/", src)
	urlHasTrailingSlash, _ := regexp.MatchString("/$", url)

	if urlHasTrailingSlash && srcHasInitialSlash {
		return strings.TrimSuffix(url, "/") + src
	} else if !urlHasTrailingSlash && !srcHasInitialSlash {
		return url + "/" + src
	} else {
		return url + src
	}
}

/**
 * Check whether a given PNG is garbage
 */
func checkImg(src string, ch chan string) {
	garbage := true
	lookup := imgUrl(src, url)
	r, err := http.Get(lookup)
	if err != nil {
		ch <- ""
		return
	}
	defer r.Body.Close()

	m, _, err := image.Decode(r.Body)
	if err != nil {
		ch <- ""
		return
	}
	bounds := m.Bounds()

Outer:
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := m.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			if a != 65535 {
				garbage = false
				break Outer
			}
		}
	}

	if garbage {
		ch <- lookup + " is garbage! Size: " + humanize.Bytes(uint64(r.ContentLength))
	} else {
		ch <- ""
	}
}

func init() {
	flag.StringVar(&url, "url", "", "The URL to check")
	flag.Parse()
}

func main() {
	if flag.Arg(0) == "version" {
		showVersion()
	}
	if url == "" {
		showUsage()
	}
	printBanner()

	fmt.Println("Checking: ", url)
	pngs := findPngs(url)
	fmt.Println("Number of pngs: ", len(pngs))

	ch := make(chan string, len(pngs))
	for _, src := range pngs {
		go checkImg(src, ch)
	}

	for i := 0; i < len(pngs); i++ {
		select {
		case r := <-ch:
			if r != "" {
				fmt.Println(r)
			}
		}
	}
}
