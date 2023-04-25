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
	vacancies := parser.ParseStartingPage(page)
	for _, vacancie := range *vacancies {
		node, err := parser.HTMLfromURL(vacancie.Url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		vacancie.HtmlNode = node
		parser.ParseVacanciePage(&vacancie)
	}
}
