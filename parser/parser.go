package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/NickBabakin/ipiad/rabbitmqgo"
	v "github.com/NickBabakin/ipiad/vacanciestructs"
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

func (p Parser) ParseStartingPage(page *html.Node) *map[string]v.VacancieMinInfo {
	var fillURLs func(*html.Node, *map[string]v.VacancieMinInfo, *regexp.Regexp)
	re, _ := regexp.Compile(`^\/vacancies\/[0-9]+$`)
	fillURLs = func(n *html.Node, vacancies *map[string]v.VacancieMinInfo, re *regexp.Regexp) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if re.MatchString(a.Val) {
						v := v.VacancieMinInfo{
							Url: "https://career.habr.com" + a.Val,
							Id:  a.Val[len(a.Val)-10:]}
						(*vacancies)[a.Val[len(a.Val)-10:]] = v

						j, err := json.Marshal(v)
						if err != nil {
							fmt.Println(err)
							return
						}
						rabbitmqgo.Send(j, "VacancieMinInfo")
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fillURLs(c, vacancies, re)
		}
	}
	vacanciesMinInfo := make(map[string]v.VacancieMinInfo)
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

func (p Parser) ParseVacanciePage(va *v.Vacancie) {
	va.Title = Query(va.HtmlNode, ".page-title__title").FirstChild.Data
	va.CompanyName = Query(va.HtmlNode, ".company_name > a").FirstChild.Data
	log.Printf("\nVacancie: \n\tId: %s\n\tTitle %s\n\tCompany name: %s\n\tUrl: %s\n\n", va.Id, va.Title, va.CompanyName, va.Url)
}
