package main

import (
	"github.com/NickBabakin/ipiad/parser"
)

func main() {
	p := parser.Parser{FirstPage: "something"}
	parser.Parse(p)
}
