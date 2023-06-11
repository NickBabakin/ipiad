package main

import (
	"fmt"

	e "github.com/NickBabakin/ipiad/res/elasticgo"
)

func main() {
	fmt.Println("I am analyzer")
	err := e.Init_elastic()
	if err != nil {
		return
	}

	e.FindDevVacancies()
	e.FindAllVacancies()

}
