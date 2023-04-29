package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/OzkrOssa/ppp-mkt-search/repository"
	"github.com/OzkrOssa/ppp-mkt-search/utils"
	"github.com/robfig/cron"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
)

type PPPoE struct {
	As            string `json:"*"`
	ID            string `json:".id"`
	CallerID      string `json:"caller-id"`
	Comment       string `json:"comment"`
	Disabled      string `json:"disabled"`
	RemoteAddress string `json:"remote-address"`
	Name          string `json:"name"`
	Profile       string `json:"profile"`
	Service       string `json:"service"`
	BTS           string `json:"bts"`
}

func main() {

	c := cron.New()
	c.AddFunc("0 0 */2 * *", mainJob)
	log.Println("cron job is started")
	c.Start()

}

func mainJob() {
	start := time.Now()
	data := utils.LoadConfig()

	var wg sync.WaitGroup
	var allSecret []map[string]string

	for _, d := range data {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			repo, err := repository.NewMikrotikRepository(p, os.Getenv("API"), os.Getenv("PASSWORD"))
			if err != nil {
				fmt.Println(err)
			}

			identity, err := repo.GetIdentity()
			if err != nil {
				fmt.Println(err)
			}

			secrets, err := repo.GetSecrets(identity["name"])
			if err != nil {
				fmt.Println(err)
			}

			allSecret = append(allSecret, secrets...)
		}(d)
	}
	wg.Wait()

	client := typesense.NewClient(
		typesense.WithServer(os.Getenv("TYPESENSE_HOST")),
		typesense.WithAPIKey(os.Getenv("TYPESENSE_API_KEY")))
	client.Collection("users").Delete()

	schema := createTypesenseSchema(client)
	client.Collections().Create(schema)

	secret := mapToStruct(allSecret)
	createTypesenseUsers(client, secret)

	elapsed := time.Since(start)
	fmt.Printf("Tiempo de ejecución: %s\n", elapsed)
}

func mapToStruct(data []map[string]string) []PPPoE {
	var finalSecrets []PPPoE
	for _, secretMap := range data {
		secretJson, err := json.Marshal(secretMap)
		if err != nil {
			fmt.Println(err)
		}

		var secrets PPPoE

		err = json.Unmarshal(secretJson, &secrets)
		if err != nil {
			fmt.Println(err)
		}

		finalSecrets = append(finalSecrets, secrets)
	}
	return finalSecrets
}

func createTypesenseSchema(client *typesense.Client) *api.CollectionSchema {
	var Facet = true
	schema := &api.CollectionSchema{
		Name: "users",
		Fields: []api.Field{
			{
				Name:  "*",
				Type:  "string",
				Facet: &Facet,
			},
			{
				Name: ".id",
				Type: "string",
			},
			{
				Name:  "caller-id",
				Type:  "string",
				Facet: &Facet,
			},
			{
				Name:  "comment",
				Type:  "string",
				Facet: &Facet,
			},
			{
				Name: "disabled",
				Type: "string",
			},
			{
				Name: "remote-address",
				Type: "string",
			},
			{
				Name:  "name",
				Type:  "string",
				Facet: &Facet,
			},
			{
				Name:  "profile",
				Type:  "string",
				Facet: &Facet,
			},
			{
				Name: "service",
				Type: "string",
			},
			{
				Name:  "bts",
				Type:  "string",
				Facet: &Facet,
			},
		},
	}

	return schema
}

func createTypesenseUsers(client *typesense.Client, data []PPPoE) {

	docChan := make(chan PPPoE)

	// Definir el número de goroutines que se van a utilizar
	numWorkers := 20

	// Iniciar las goroutines
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for doc := range docChan {
				client.Collection("users").Documents().Create(doc)
			}
		}()
	}

	for _, s := range data {
		docChan <- PPPoE{
			ID:            s.ID,
			CallerID:      s.CallerID,
			Comment:       s.Comment,
			Disabled:      s.Disabled,
			RemoteAddress: s.RemoteAddress,
			Name:          s.Name,
			Profile:       s.Profile,
			Service:       s.Service,
			BTS:           s.BTS,
		}
	}

	// Cerrar el canal y esperar a que todas las goroutines finalicen
	close(docChan)
	wg.Wait()
}
