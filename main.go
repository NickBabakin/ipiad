package main

import (
	"sync"

	p "github.com/NickBabakin/ipiad/parser"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(2)

	go p.ParseStartingPage("https://career.habr.com/vacancies", &wg)
	go p.ParseVacancies(&wg)

	wg.Wait()
}
