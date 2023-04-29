package main

import (
	"fmt"

	p "github.com/NickBabakin/ipiad/parser"
)

func main() {
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
