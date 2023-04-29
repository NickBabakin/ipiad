package parser

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/andybalholm/cascadia"

	"golang.org/x/net/html"
)

func Query(n *html.Node, query string) *html.Node {
	sel, err := cascadia.Parse(query)
	if err != nil {
		return &html.Node{}
	}
	return cascadia.Query(n, sel)
}

type Parser struct{}

type Vacancie struct {
	Url         string
	Title       string
	CompanyName string
	Id          string
	HtmlNode    *html.Node
}

type VacancieMinInfo struct {
	Url string `json:"url"`
	Id  string `json:"id"`
}

func (p Parser) ParseStartingPage(page *html.Node) *map[string]VacancieMinInfo {
	var fillURLs func(*html.Node, *map[string]VacancieMinInfo, *regexp.Regexp)
	re, _ := regexp.Compile(`^\/vacancies\/[0-9]+$`)
	fillURLs = func(n *html.Node, vacancies *map[string]VacancieMinInfo, re *regexp.Regexp) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if re.MatchString(a.Val) {
						(*vacancies)[a.Val[len(a.Val)-10:]] = VacancieMinInfo{
							Url: "https://career.habr.com" + a.Val,
							Id:  a.Val[len(a.Val)-10:]}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fillURLs(c, vacancies, re)
		}
	}
	vacanciesMinInfo := make(map[string]VacancieMinInfo)
	fillURLs(page, &vacanciesMinInfo, re)
	return &vacanciesMinInfo
}

func (p Parser) HTMLfromURL(url string) (*html.Node, error) {
	log.Printf("Requesting GET %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("request status is " + strconv.Itoa(res.StatusCode) + ". Unable to fetch data")
	}
	page, err := html.Parse(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	return page, nil
}

func (p Parser) ParseVacanciePage(v *Vacancie) {
	v.Title = Query(v.HtmlNode, ".page-title__title").FirstChild.Data
	v.CompanyName = Query(v.HtmlNode, ".company_name > a").FirstChild.Data
	log.Printf("\nVacancie: \n\tId: %s\n\tTitle %s\n\tCompany name: %s\n\tUrl: %s\n\n", v.Id, v.Title, v.CompanyName, v.Url)
}
