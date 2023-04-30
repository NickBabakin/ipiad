package main

import (
	"sync"

	p "github.com/NickBabakin/ipiad/parser"
	"github.com/NickBabakin/ipiad/rabbitmqgo"
)

var vacancieMinInfoStr string = "VacancieMinInfo"

func main() {
	ch := make(chan []byte, 100)
	var wg sync.WaitGroup
	wg.Add(2)

	go rabbitmqgo.Receive(vacancieMinInfoStr, &wg, ch)
	go p.ParseVacancies(ch, &wg)

	p.ParseStartingPage("https://career.habr.com/vacancies")

	wg.Wait()
}
