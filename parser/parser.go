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
	//rabbitmqgo.Send([]byte("stop"), vacancieMinInfoStr)
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
	defer rabbit.Conn.Close()
	defer rabbit.Ch.Close()

	vmis := make(chan *amqp.Delivery, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go rabbitmqgo.Receive(vacancieMinInfoStr, &wg, vmis, rabbit)

	for vmi := range vmis {
		if string(vmi.Body) == "stop" {
			err := vmi.Ack(false)
			if err != nil {
				log.Println(err)
			}
			//rabbitmqgo.Send([]byte("stop"), vacancieFullInfoStr)
			break
		}

		wg.Add(1)
		go func(vmi *amqp.Delivery) {
			defer wg.Done()
			err := vmi.Ack(false)
			if err != nil {
				log.Println(err)
			}
			var va v.VacancieMinInfo
			log.Printf("Received a message: %s\n", vmi.Body)
			json.Unmarshal(vmi.Body, &va)
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
			log.Printf("\nVacancie: \n\tId: %s\n\tTitle %s\n\tCompany name: %s\n\tUrl: %s\n\n", vfi.Id, vfi.Title, vfi.CompanyName, vfi.Url)

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
	defer rabbit.Conn.Close()
	defer rabbit.Ch.Close()

	vfis := make(chan *amqp.Delivery, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go rabbitmqgo.Receive(vacancieFullInfoStr, &wg, vfis, rabbit)

	for vfi := range vfis {
		if string(vfi.Body) == "stop" {
			err := vfi.Ack(false)
			if err != nil {
				log.Println(err)
			}
			break
		}

		wg.Add(1)
		go func(vfi *amqp.Delivery) {
			defer wg.Done()
			log.Printf("Got full info %s\n", vfi.Body)
			err := vfi.Ack(false)
			if err != nil {
				log.Println(err)
			}
		}(vfi)
	}

	wg.Wait()
	log.Printf(" All VacancieFullInfo processed\n")
}
