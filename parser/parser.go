package parser

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type Parser struct{}

func getVacanciesURLs(n *html.Node) *[]string {
	var fillURLs func(*html.Node, *[]string, *regexp.Regexp)
	re, _ := regexp.Compile(`^\/vacancies\/[0-9]+$`)
	fillURLs = func(n *html.Node, vacanciesURLs *[]string, re *regexp.Regexp) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if re.MatchString(a.Val) {
						*vacanciesURLs = append(*vacanciesURLs, "https://career.habr.com"+a.Val)
						//TODO: append unique only
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fillURLs(c, vacanciesURLs, re)
		}
	}
	vacanciesURLs := make([]string, 0, 512)
	fillURLs(n, &vacanciesURLs, re)
	return &vacanciesURLs
}

func (p Parser) HTMLfromURL(url string) *html.Node {
	res, err := http.Get(url)
	//TODO: check for http error codes
	if err != nil {
		log.Fatal(err)
	}
	html_data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	html_string := string(html_data)
	fmt.Printf(html_string, "\nhey\n\n\n\n\n\n\n\n\n\n\n")
	doc, err := html.Parse(strings.NewReader(html_string))
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func (p Parser) ParseStartingHTML(doc *html.Node) *[]string {
	vacanciesURLs := getVacanciesURLs(doc)
	return vacanciesURLs
}
