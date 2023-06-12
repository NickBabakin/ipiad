package elasticgo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch"
)

var Cfg elasticsearch.Config = elasticsearch.Config{
	Addresses: []string{
		//"http://es-container:9200", // with docker
		"http://localhost:9200", // without docker
	},
}

var es *elasticsearch.Client

func Init_elastic() error {

	es_client, err := elasticsearch.NewClient(Cfg)
	if err != nil {
		log.Printf("Error creating the client: %s", err)
		return err
	}
	es = es_client

	res, err := es.Info()
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return err
	}
	defer res.Body.Close()
	return nil
}

func IndexVacancie(vacancie string, id string, opType string) error {

	res, err := es.Index(
		"vacancies",                         // Index name
		strings.NewReader(string(vacancie)), // Document body
		es.Index.WithDocumentID(id),         // Document ID
		es.Index.WithRefresh("true"),        // Refresh
		es.Index.WithOpType(opType),         // optype "create" fails when there already exists doc with given id,
		es.Index.WithPretty(),               // optype "index" allows updating
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 409 {
		return errors.New("Doc already exists and was not updated")
	}
	log.Printf("Successfully indexing vacancie : %s", vacancie)

	log.Println(res)
	return nil
}

func GetVacancieById(id string) (string, error) {
	res, err := es.Get("vacancies", id)
	if err != nil {
		fmt.Println("error getting doc : ", err)
		return "", err
	}
	defer res.Body.Close()
	vacancie_res_str, _ := io.ReadAll(res.Body)
	return string(vacancie_res_str), nil
}

func search(query map[string]interface{}) map[string]interface{} {
	var (
		r map[string]interface{}
	)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("vacancies"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)

	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)

	return r
}

func SearchMtsDevVacancies() {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"company_name": map[string]interface{}{
								"query":     "МТС",
								"fuzziness": "AUTO",
							},
						},
					},
				},
				"should": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"title": map[string]interface{}{
								"query":     "Разработчик",
								"fuzziness": "AUTO",
							},
						},
					},
					{
						"match": map[string]interface{}{
							"title": map[string]interface{}{
								"query":     "Developer",
								"fuzziness": "AUTO",
							},
						},
					},
				},
			},
		},
	}

	search(query)
}

func SearchAllVacancies() {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	search(query)
}

type Profession int

const (
	Developer Profession = iota
	Analyst
	Architect
	Other
)

type VacancieProfession struct {
	Id         string
	Profession Profession
	Title      string
}

var ProffessionToString = map[Profession]string{
	Developer: "Developer",
	Analyst:   "Analyst",
	Architect: "Architect",
	Other:     "Other",
}

var developerEnStr string = ".*eveloper.*"
var developerRuStr string = ".*азработчик.*"
var analystEnStr string = ".*nalyst.*"
var analystRuStr string = ".*налитик.*"
var architectEnStr string = ".*rchitect.*"
var architectRuStr string = ".*рхитектор.*"

func newVacancieProfessionSlice(profession Profession, source map[string]interface{}) []VacancieProfession {
	vacancieProfessionSlice := make([]VacancieProfession, 0, 500)

	for _, hit := range source["hits"].(map[string]interface{})["hits"].([]interface{}) {
		id := fmt.Sprintf("%v", hit.(map[string]interface{})["_id"])
		title := fmt.Sprintf("%v", hit.(map[string]interface{})["_source"].(map[string]interface{})["title"])
		//		log.Printf(" * ID=%s, %s", id, title)
		vacancieProfessionSlice = append(vacancieProfessionSlice, VacancieProfession{
			Id:         id,
			Title:      title,
			Profession: profession,
		})
	}
	return vacancieProfessionSlice
}

func SearchProfessionVacancies(profession Profession) []VacancieProfession {
	var profRuStr string
	var profEnStr string

	switch profession {
	case Developer:
		profRuStr = developerRuStr
		profEnStr = developerEnStr
	case Analyst:
		profRuStr = analystRuStr
		profEnStr = analystEnStr
	case Architect:
		profRuStr = architectRuStr
		profEnStr = architectEnStr
	case Other:
		return SearchOtherVacancies()
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": profEnStr,
							},
						},
					},
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": profRuStr,
							},
						},
					},
				},
			},
		},
		"size": 500,
	}

	res := search(query)

	return newVacancieProfessionSlice(profession, res)
}

func SearchOtherVacancies() []VacancieProfession {

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must_not": []map[string]interface{}{
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": developerRuStr,
							},
						},
					},
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": developerEnStr,
							},
						},
					},
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": analystRuStr,
							},
						},
					},
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": analystEnStr,
							},
						},
					},
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": architectRuStr,
							},
						},
					},
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": architectEnStr,
							},
						},
					},
				},
			},
		},
		"size": 500,
	}

	res := search(query)

	return newVacancieProfessionSlice(Profession(Other), res)
}
