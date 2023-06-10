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
	res, err := e.GetVacancieById("1000119848")
	if err != nil {
		fmt.Println("error gerring doc : ", err)
	}
	fmt.Println(res)

	e.FindAllvacancies()

}
