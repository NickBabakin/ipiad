package main

import (
	"encoding/json"
	"fmt"
	"time"

	p "github.com/NickBabakin/ipiad/parser"
	"github.com/NickBabakin/ipiad/rabbitmqgo"
)

func main() {

	v := p.VacancieMinInfo{
		Url: "vacancieMinInfo.Url",
		Id:  "vacancieMinInfo.Id"}
	j, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
		return
	}
	rabbitmqgo.Send(j)
	time.Sleep(2 * time.Second)
	rabbitmqgo.Receive()

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
		vacancie := p.Vacancie{
			Url:      vacancieMinInfo.Url,
			Id:       vacancieMinInfo.Id,
			HtmlNode: node,
		}
		parser.ParseVacanciePage(&vacancie)
	}
}
