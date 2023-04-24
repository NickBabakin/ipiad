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
	vacanciesURLs := parser.ParseStartingPage(page)
	for _, vacancieURL := range *vacanciesURLs {
		vacancie, err := parser.HTMLfromURL(vacancieURL)
		if err != nil {
			fmt.Println(err)
			continue
		}
		parser.ParseVacanciePage(vacancie)
	}
}
