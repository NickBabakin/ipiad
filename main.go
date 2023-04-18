package main

import (
	"fmt"

	p "github.com/NickBabakin/ipiad/parser"
)

func main() {
	parser := p.Parser{}
	html := parser.HTMLfromURL("https://career.habr.com/vacancies")
	fmt.Printf("\n\n\n\n NOW TO LINK \n\n\n\n")
	parser.ParseHTML(html)
}
