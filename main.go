package main

import (
	"io"
	"log"
	"os"
	"sync"
	"time"

	e "github.com/NickBabakin/ipiad/elasticgo"
	p "github.com/NickBabakin/ipiad/parser"
)

func work() {

	for {
		err := e.Init_elastic()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 5)
	}

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

func main() {
	work()
	//e.F_elastic()

}
