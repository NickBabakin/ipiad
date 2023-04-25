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

func getVacancies(n *html.Node) *[]Vacancie {
	var fillURLs func(*html.Node, *[]Vacancie, *regexp.Regexp)
	re, _ := regexp.Compile(`^\/vacancies\/[0-9]+$`)
	fillURLs = func(n *html.Node, vacancies *[]Vacancie, re *regexp.Regexp) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if re.MatchString(a.Val) {
						*vacancies = append(*vacancies, Vacancie{
							Url: "https://career.habr.com" + a.Val,
							Id:  a.Val[len(a.Val)-10:],
						})
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fillURLs(c, vacancies, re)
		}
	}
	vacancies := make([]Vacancie, 0, 512)
	fillURLs(n, &vacancies, re)
	return &vacancies
}

func (p Parser) ParseStartingPage(page *html.Node) *[]Vacancie {
	vacancies := getVacancies(page)
	return vacancies
}

func (p Parser) HTMLfromURL(url string) (*html.Node, error) {
	res, err := http.Get(url)
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
	page, err := html.Parse(strings.NewReader(html_string))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return page, nil
}

func (p Parser) ParseVacanciePage(v *Vacancie) {
	v.Title = Query(v.HtmlNode, ".page-title__title").FirstChild.Data
	v.CompanyName = Query(v.HtmlNode, ".company_name > a").FirstChild.Data
	fmt.Printf("\nId: %s\nTitle %s\nCompany name: %s\nURL: %s\n", v.Id, v.Title, v.CompanyName, v.Url)
}
