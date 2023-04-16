package parser

import "fmt"

type Parser struct {
	FirstPage string
}

func Parse(p Parser) {
	fmt.Printf("Hello, Worlds! %s", p.FirstPage)
}
