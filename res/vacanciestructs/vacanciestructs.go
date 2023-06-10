package vacanciestructs

import "golang.org/x/net/html"

type Vacancie struct {
	Url         string
	Title       string
	CompanyName string
	Id          string
	HtmlNode    *html.Node
}

type VacancieMinInfo struct {
	Url string `json:"url"`
	Id  string `json:"id"`
}

type VacancieFullInfo struct {
	Url         string `json:"url"`
	Id          string `json:"id"`
	Title       string `json:"title"`
	CompanyName string `json:"company_name"`
}
