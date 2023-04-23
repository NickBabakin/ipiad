package main

import (
	p "github.com/NickBabakin/ipiad/parser"
)

func main() {
	parser := p.Parser{}
	doc := parser.HTMLfromURL("https://career.habr.com/vacancies")
	parser.ParseStartingHTML(doc)
}
