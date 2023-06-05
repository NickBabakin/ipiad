package main

import (
	"io"
	"log"
	"os"
	"sync"

	p "github.com/NickBabakin/ipiad/parser"
)

func main() {

	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
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
