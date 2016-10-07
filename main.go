package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"net/http"
	"regexp"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
)

var url string

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
func checkImg(src string) error {
	garbage := true
	r, err := http.Get(imgUrl(src, url))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	m, _, err := image.Decode(r.Body)
	if err != nil {
		return err
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
		fmt.Println(src, " is garbage! Content-Length: ", r.ContentLength)
	}

	return nil
}

func init() {
	flag.StringVar(&url, "url", "", "The URL to check")
	flag.Parse()
}

func main() {
	printBanner()

	fmt.Println("Checking: ", url)
	pngs := findPngs(url)
	fmt.Println("Number of pngs: ", len(pngs))

	for _, src := range pngs {
		err := checkImg(src)
		if err != nil {
			fmt.Println(err)
		}
	}
}
