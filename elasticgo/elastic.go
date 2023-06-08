package elasticgo

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	v "github.com/NickBabakin/ipiad/vacanciestructs"
	"github.com/elastic/go-elasticsearch"
)

var Cfg elasticsearch.Config = elasticsearch.Config{
	Addresses: []string{
		"http://localhost:9200",
	},
}

var Es *elasticsearch.Client

func Init_elastic() error {

	Es_client, err := elasticsearch.NewClient(Cfg)
	if err != nil {
		log.Printf("Error creating the client: %s", err)
		return err
	}

	Es = Es_client

	res, err := Es.Info()
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return err
	}
	fmt.Println(res)

	defer res.Body.Close()

	log.Println(res)
	return nil
}

func IndexVacancie(vacancie v.VacancieFullInfo) {
	vacancie_str, _ := json.Marshal(vacancie)

	log.Printf(" indexing vacancie : %s", vacancie_str)
	res, err := Es.Index(
		"vacancies",                             // Index name
		strings.NewReader(string(vacancie_str)), // Document body
		Es.Index.WithDocumentID(vacancie.Id),    // Document ID
		Es.Index.WithRefresh("true"),            // Refresh
		Es.Index.WithPretty(),
	)
	if err != nil {
		log.Printf("Error indexing vacancie : %s", err)
	}
	defer res.Body.Close()

	log.Println(res)
}
