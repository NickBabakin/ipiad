package main

import (
	"fmt"

	p "github.com/NickBabakin/ipiad/parser"
)

func main() {
	parser := p.Parser{}
	doc := parser.HTMLfromURL("https://career.habr.com/vacancies")
	vacanciesURLs := parser.ParseStartingHTML(doc)
	//for _, s := range *vacanciesURLs {
	//	fmt.Println(s)
	//}
	//fmt.Println(len(*vacanciesURLs))
	vacancie := parser.HTMLfromURL((*vacanciesURLs)[0])
	fmt.Println(vacancie)
}
