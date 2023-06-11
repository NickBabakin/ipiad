package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	e "github.com/NickBabakin/ipiad/res/elasticgo"
	"github.com/NickBabakin/ipiad/res/rabbitmqgo"
	v "github.com/NickBabakin/ipiad/res/vacanciestructs"
	"github.com/andybalholm/cascadia"
	amqp "github.com/rabbitmq/amqp091-go"

	"golang.org/x/net/html"
)

var vacancieMinInfoStr string = "VacancieMinInfo"
var vacancieFullInfoStr string = "VacancieFullInfo"
var habrStr string = "https://career.habr.com/vacancies?type=all&page="

func Query(n *html.Node, query string) *html.Node {
	sel, err := cascadia.Parse(query)
	if err != nil {
		return &html.Node{}
	}
	return cascadia.Query(n, sel)
}

func ParseStartingPage(wg_ext *sync.WaitGroup) {
	defer wg_ext.Done()

	var pages [10]*html.Node
	for i := 0; i < 10; i++ {
		page, err := HTMLfromURL(habrStr + strconv.Itoa(i))
		if err != nil {
			fmt.Println(err)
			return
		}
		pages[i] = page
	}

	var fillURLs func(*html.Node, *regexp.Regexp)
	re, _ := regexp.Compile(`^\/vacancies\/[0-9]+$`)
	fillURLs = func(n *html.Node, re *regexp.Regexp) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if re.MatchString(a.Val) {
						id := a.Val[len(a.Val)-10:]
						vacancie := v.VacancieMinInfo{
							Url: "https://career.habr.com" + a.Val,
							Id:  id}

						vacancieJson, err := json.Marshal(vacancie)
						if err != nil {
							fmt.Println(err)
							return
						}
						err = e.IndexVacancie(string(vacancieJson), vacancie.Id, "create")
						if err != nil {
							log.Printf("Tried to index %s : %s\n", vacancie.Id, err)
							break
						}
						rabbitmqgo.Send(vacancieJson, vacancieMinInfoStr)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fillURLs(c, re)
		}
	}

	for _, page := range pages {
		if page == nil {
			break
		}
		fillURLs(page, re)
	}
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
	va.Date = Query(va.HtmlNode, "div.vacancy-header__date > span > time").Attr[1].Val[0:10]
	return v.VacancieFullInfo{
		Id:          va.Id,
		Title:       va.Title,
		CompanyName: va.CompanyName,
		Url:         va.Url,
		Date:        va.Date,
	}
}

func ParseVacancies(wg_ext *sync.WaitGroup) {
	defer wg_ext.Done()

	rabbit := rabbitmqgo.InitRabbit()

	var wg sync.WaitGroup
	wg.Add(1)
	vmis := rabbitmqgo.Receive(vacancieMinInfoStr, &wg, rabbit)

	for vmi := range vmis {
		wg.Add(1)
		go func(vmi_e amqp.Delivery) {
			defer wg.Done()
			var va v.VacancieMinInfo
			log.Printf("ParseVacancies: %s\n", vmi_e.Body)
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

			j, err := json.Marshal(vfi)
			if err != nil {
				log.Println(err)
				return
			}
			rabbitmqgo.Send(j, vacancieFullInfoStr)
			err = vmi_e.Ack(false)
			if err != nil {
				log.Println("ACK error" + err.Error())
			}
		}(vmi)

	}
	wg.Wait()
	log.Printf(" All VacancieMinInfo processed\n")
}

func SaveVacancies(wg_ext *sync.WaitGroup) {
	defer wg_ext.Done()

	rabbit := rabbitmqgo.InitRabbit()

	var wg sync.WaitGroup
	wg.Add(1)
	vfis := rabbitmqgo.Receive(vacancieFullInfoStr, &wg, rabbit)

	for vfi := range vfis {
		wg.Add(1)
		go func(vfi_e amqp.Delivery) {
			defer wg.Done()
			log.Printf("SaveVacancies %s\n", vfi_e.Body)
			var va v.VacancieFullInfo
			json.Unmarshal(vfi_e.Body, &va)
			e.IndexVacancie(string(vfi_e.Body), va.Id, "index")
			err := vfi_e.Ack(false)
			if err != nil {
				log.Println("ACK error" + err.Error())
			}
		}(vfi)
	}

	wg.Wait()
	log.Printf(" All VacancieFullInfo processed\n")
}

func Work() {
	for {
		err := e.Init_elastic()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 5)
	}

	logFile, err := os.OpenFile("logs/log.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	var wg sync.WaitGroup
	wg.Add(3)

	go ParseStartingPage(&wg)
	go ParseVacancies(&wg)
	go SaveVacancies(&wg)

	wg.Wait()
}
