package main

import (
	"sync"

	p "github.com/NickBabakin/ipiad/parser"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(3)

	go p.ParseStartingPage("https://career.habr.com/vacancies", &wg)
	go p.ParseVacancies(&wg)
	go p.SaveVacancies(&wg)

	wg.Wait()
}
