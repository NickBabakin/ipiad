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
		"http://localhost:9200",
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
	fmt.Println(res)

	defer res.Body.Close()

	log.Println(res)
	return nil
}

/* func IndexVacancie(vacancie string, id string) {

	log.Printf("Successfully indexing vacancie : %s", vacancie)
	res, err := es.Index(
		"vacancies",                         // Index name
		strings.NewReader(string(vacancie)), // Document body
		es.Index.WithDocumentID(id),         // Document ID
		es.Index.WithRefresh("true"),        // Refresh
		es.Index.WithPretty(),
	)
	if err != nil {
		log.Printf("Error indexing vacancie : %s", err)
	}
	defer res.Body.Close()

	log.Println(res)
} */

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

func FindAllvacancies() {
	var (
		r map[string]interface{}
		//wg sync.WaitGroup
	)

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"title": map[string]interface{}{
					"query":     "Разработчик",
					"fuzziness": "AUTO",
				},
			},
		},
	}
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

	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

	log.Println(strings.Repeat("=", 37))

}
