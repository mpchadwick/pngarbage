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
		garbage := true
		lookup := imgUrl(src, url)
		r, err := http.Get(lookup)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer r.Body.Close()

		m, _, err := image.Decode(r.Body)
		if err != nil {
			continue
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
	}
}
