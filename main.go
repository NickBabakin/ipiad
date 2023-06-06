package main

import (
	"io"
	"log"
	"os"
	"sync"

	p "github.com/NickBabakin/ipiad/parser"
	"github.com/elastic/go-elasticsearch/v8"
)

func work() {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	var wg sync.WaitGroup
	wg.Add(3)

	go p.ParseStartingPage("https://career.habr.com/vacancies", &wg)
	go p.ParseVacancies(&wg)
	go p.SaveVacancies(&wg)

	wg.Wait()
}

func f_elastic() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()
	log.Println(res)
}

func main() {
	//work()
	f_elastic()

}
