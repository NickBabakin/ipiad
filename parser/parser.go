package parser

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type Parser struct{}

func getVacanciesURLs(n *html.Node) []string {
	var fillURLs func(*html.Node, *[]string)
	fillURLs = func(n *html.Node, vacanciesURLs *[]string) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					//fmt.Println(a.Val)
					*vacanciesURLs = append(*vacanciesURLs, a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fillURLs(c, vacanciesURLs)
		}
		//fmt.Println(vacanciesURLs)
	}
	vacanciesURLs := make([]string, 0, 512)
	fillURLs(n, &vacanciesURLs)
	return vacanciesURLs
}

func (p Parser) HTMLfromURL(url string) string {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	html_data, err := io.ReadAll(res.Body)
	html := string(html_data)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	return html
}

func (p Parser) ParseHTML(html_string string) {
	doc, err := html.Parse(strings.NewReader(html_string))
	if err != nil {
		log.Fatal(err)
	}
	vacanciesURLs := getVacanciesURLs(doc)
	for _, s := range vacanciesURLs {
		fmt.Println(s)
	}
	fmt.Println(len(vacanciesURLs), "  ", cap(vacanciesURLs))
}
