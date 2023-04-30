package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"

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

func ParseStartingPage(habr_str string) *map[string]v.VacancieMinInfo {

	page, err := HTMLfromURL(habr_str)
	if err != nil {
		fmt.Println(err)
		return nil
	}

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

func HTMLfromURL(url string) (*html.Node, error) {
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

func ParseVacanciePage(va *v.Vacancie) {
	va.Title = Query(va.HtmlNode, ".page-title__title").FirstChild.Data
	va.CompanyName = Query(va.HtmlNode, ".company_name > a").FirstChild.Data
	log.Printf("\nVacancie: \n\tId: %s\n\tTitle %s\n\tCompany name: %s\n\tUrl: %s\n\n", va.Id, va.Title, va.CompanyName, va.Url)
}

func ParseVacancies(chv chan []byte, wg_ext *sync.WaitGroup) {
	defer wg_ext.Done()
	for msg := range chv {
		var va v.VacancieMinInfo
		log.Printf("Received a message: %s\n", msg)
		json.Unmarshal(msg, &va)
		log.Printf("\nVacancie unmarshaled: \n\tId: %s\n\tUrl: %s\n\n", va.Id, va.Url)
		node, err := HTMLfromURL(va.Url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		vacancie := v.Vacancie{
			Url:      va.Url,
			Id:       va.Id,
			HtmlNode: node,
		}
		ParseVacanciePage(&vacancie)
	}
}
