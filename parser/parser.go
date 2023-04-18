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
	//fmt.Println(res.Status, "\n\n\n", html)
	return html
}

func (p Parser) ParseHTML(html_string string) {
	fmt.Printf("Hello, World!\n")
	doc, err := html.Parse(strings.NewReader(html_string))
	if err != nil {
		log.Fatal(err)
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					fmt.Println(a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}
