package main

import (
	"fmt"
	"sync"

	p "github.com/NickBabakin/ipiad/parser"
	"github.com/NickBabakin/ipiad/rabbitmqgo"
	v "github.com/NickBabakin/ipiad/vacanciestructs"
)

var vacancieMinInfoStr string = "VacancieMinInfo"

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go rabbitmqgo.Receive(vacancieMinInfoStr, &wg)

	parser := p.Parser{}
	page, err := parser.HTMLfromURL("https://career.habr.com/vacancies")
	if err != nil {
		fmt.Println(err)
		return
	}
	vacanciesMinInfo := parser.ParseStartingPage(page)

	for _, vacancieMinInfo := range *vacanciesMinInfo {
		node, err := parser.HTMLfromURL(vacancieMinInfo.Url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		vacancie := v.Vacancie{
			Url:      vacancieMinInfo.Url,
			Id:       vacancieMinInfo.Id,
			HtmlNode: node,
		}
		parser.ParseVacanciePage(&vacancie)
	}
	wg.Wait()
}
