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

var vacancieMinInfoStr string = "VacancieMinInfo"

func Query(n *html.Node, query string) *html.Node {
	sel, err := cascadia.Parse(query)
	if err != nil {
		return &html.Node{}
	}
	return cascadia.Query(n, sel)
}

var i int = 0

func ParseStartingPage(habr_str string, wg_ext *sync.WaitGroup) {
	defer wg_ext.Done()

	page, err := HTMLfromURL(habr_str)
	if err != nil {
		fmt.Println(err)
		return
	}

	var fillURLs func(*html.Node, *regexp.Regexp)
	re, _ := regexp.Compile(`^\/vacancies\/[0-9]+$`)
	fillURLs = func(n *html.Node, re *regexp.Regexp) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if re.MatchString(a.Val) {
						i++
						if i > 5 {
							return
						}
						v := v.VacancieMinInfo{
							Url: "https://career.habr.com" + a.Val,
							Id:  a.Val[len(a.Val)-10:]}

						j, err := json.Marshal(v)
						if err != nil {
							fmt.Println(err)
							return
						}
						rabbitmqgo.Send(j, vacancieMinInfoStr)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fillURLs(c, re)
		}
	}
	fillURLs(page, re)
	rabbitmqgo.Send([]byte("stop"), vacancieMinInfoStr)
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

func ParseVacancies(wg_ext *sync.WaitGroup) {
	defer wg_ext.Done()

	vmis := make(chan []byte, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go rabbitmqgo.Receive(vacancieMinInfoStr, &wg, vmis)

	for vmi := range vmis {
		if string(vmi) == "stop" {
			break
		}

		wg.Add(1)
		go func(vmi_c []byte) {
			defer wg.Done()
			var va v.VacancieMinInfo
			log.Printf("Received a message: %s\n", vmi_c)
			json.Unmarshal(vmi_c, &va)
			node, err := HTMLfromURL(va.Url)
			if err != nil {
				fmt.Println(err)
				return
			}
			vacancie := v.Vacancie{
				Url:      va.Url,
				Id:       va.Id,
				HtmlNode: node,
			}
			ParseVacanciePage(&vacancie)
		}(vmi)

	}
	log.Printf(" All VacancieMinInfo processed\n")
	wg.Wait()
}
