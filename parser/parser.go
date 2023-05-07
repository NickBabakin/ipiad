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
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/exp/slices"

	"golang.org/x/net/html"
)

var vacancieMinInfoStr string = "VacancieMinInfo"
var vacancieFullInfoStr string = "VacancieFullInfo"

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

	ids := []string{}

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
						id := a.Val[len(a.Val)-10:]
						if slices.Contains(ids, id) {
							return
						} else {
							ids = append(ids, id)
						}
						i++
						if i > 20 {
							return
						}
						v := v.VacancieMinInfo{
							Url: "https://career.habr.com" + a.Val,
							Id:  id}

						j, err := json.Marshal(v)
						if err != nil {
							fmt.Println(err)
							return
						}
						//time.Sleep(time.Second * 2)
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
}

func HTMLfromURL(url string) (*html.Node, error) {
	//log.Printf("Requesting GET %s\n", url)
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

func ParseVacanciePage(va *v.Vacancie) v.VacancieFullInfo {
	va.Title = Query(va.HtmlNode, ".page-title__title").FirstChild.Data
	va.CompanyName = Query(va.HtmlNode, ".company_name > a").FirstChild.Data
	return v.VacancieFullInfo{
		Id:          va.Id,
		Title:       va.Title,
		CompanyName: va.CompanyName,
		Url:         va.Url,
	}
}

func ParseVacancies(wg_ext *sync.WaitGroup) {
	defer wg_ext.Done()

	rabbit := rabbitmqgo.InitRabbit()

	vmis := make(chan *amqp.Delivery, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go rabbitmqgo.Receive(vacancieMinInfoStr, &wg, vmis, rabbit)

	for vmi := range vmis {
		wg.Add(1)
		go func(vmi_e *amqp.Delivery) {
			defer wg.Done()
			err := vmi_e.Ack(false)
			if err != nil {
				log.Println("ACK error" + err.Error())
			}
			var va v.VacancieMinInfo
			//log.Printf("ParseVacancies: %s\n", vmi_e.Body)
			json.Unmarshal(vmi_e.Body, &va)
			node, err := HTMLfromURL(va.Url)
			if err != nil {
				log.Println(err)
				return
			}
			vacancie := v.Vacancie{
				Url:      va.Url,
				Id:       va.Id,
				HtmlNode: node,
			}
			vfi := ParseVacanciePage(&vacancie)
			//log.Printf("\nVacancie: \n\tId: %s\n\tTitle %s\n\tCompany name: %s\n\tUrl: %s\n\n", vfi.Id, vfi.Title, vfi.CompanyName, vfi.Url)

			j, err := json.Marshal(vfi)
			if err != nil {
				log.Println(err)
				return
			}
			rabbitmqgo.Send(j, vacancieFullInfoStr)
		}(vmi)

	}
	wg.Wait()
	log.Printf(" All VacancieMinInfo processed\n")
}

func SaveVacancies(wg_ext *sync.WaitGroup) {
	defer wg_ext.Done()

	rabbit := rabbitmqgo.InitRabbit()

	vfis := make(chan *amqp.Delivery, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go rabbitmqgo.Receive(vacancieFullInfoStr, &wg, vfis, rabbit)

	for vfi := range vfis {
		wg.Add(1)
		go func(vfi_e *amqp.Delivery) {
			defer wg.Done()
			err := vfi_e.Ack(false)
			if err != nil {
				log.Println("ACK error" + err.Error())
			}
			log.Printf("SaveVacancies %s\n", vfi_e.Body)
		}(vfi)
	}

	wg.Wait()
	log.Printf(" All VacancieFullInfo processed\n")
}
