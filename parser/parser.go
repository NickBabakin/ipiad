package parser

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/andybalholm/cascadia"

	"golang.org/x/net/html"
)

type Parser struct{}

func Query(n *html.Node, query string) *html.Node {
	sel, err := cascadia.Parse(query)
	if err != nil {
		return &html.Node{}
	}
	return cascadia.Query(n, sel)
}

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

func (p Parser) HTMLfromURL(url string) (*html.Node, error) {
	res, err := http.Get(url)
	//TODO: check for http error codes
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("request status is " + strconv.Itoa(res.StatusCode) + ". Unable to fetch data")
	}
	html_data, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	html_string := string(html_data)
	//fmt.Printf(html_string, "\nhey\n\n\n\n\n\n\n\n\n\n\n")
	page, err := html.Parse(strings.NewReader(html_string))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return page, nil
}

func (p Parser) ParseStartingPage(page *html.Node) *[]string {
	vacanciesURLs := getVacanciesURLs(page)
	return vacanciesURLs
}

func (p Parser) ParseVacanciePage(page *html.Node) {
	title := Query(page, ".page-title__title").FirstChild.Data
	company_name := Query(page, ".company_name > a").FirstChild.Data

	fmt.Println("\n", title)
	fmt.Println("", company_name)
}
