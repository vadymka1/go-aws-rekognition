package services

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"time"

	//"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/rekognition"
	//"github.com/elastic/go-elasticsearch"
	elastic "gopkg.in/olivere/elastic.v7"
)

const (
	index  = "images"
	mapping = `
		{
			"settings":{
				"number_of_shards":2,
				"number_of_replicas":1
			},
			"mappings":{
				"image" {
					"properties":{
						"name":{
							"type":"string"
						}
						"id":{
							"type":"long"
						}
						"label":{
							"type":"nested"
						}
						"properties":{
							"name":{
								"type":"string"
							}
							"confidence":{
								"type":"float"
							}
						}
					}
				}			
			}
		}
	`
)

//var client *elastic.Client
var ctx = context.Background()

type ImageLabels struct {
	ID 		   uint     `json:"id"`
	Name 	   string     `json:"name"`
	Label      []*rekognition.Label   `json:"label"`
}

type Label struct {
	Name 	   string `json:"name"`
	Confidence float64 `json:"confidence"`
}

func init() {
	checkIndex()
}

func (i *ImageLabels) String() string {
	sum := fmt.Sprintf("\n\tID: Image: %s, Label (%d):", i.Name, len(i.Label))
	for _, lb := range i.Label {
		sum = fmt.Sprintf("%s\n\t\t[ Name: %s, Level: %d ]", sum, lb.Name, lb.Confidence)
	}
	return sum
}

func checkIndex() {
	client, err := elastic.NewClient(
		elastic.SetSniff(true),
		elastic.SetURL("http://localhost:9200"),
		elastic.SetHealthcheckInterval(5*time.Second),)
	if err != nil {
		// (Bad Request): Failed to parse content to map if mapping bad
		fmt.Println("elastic.NewClient() ERROR: %v", err)
		log.Fatalf("quiting connection..")
	} else {
		// Print client information
		fmt.Println("client:", client)
		fmt.Println("client TYPE:", reflect.TypeOf(client), "\n")
	}

	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		fmt.Println("problem with index: ", err)
	}

	if !exists {
		createIndex, err := client.CreateIndex(index).BodyString(mapping).Do(ctx)
		if err != nil {
			fmt.Println("problem with creation index: ", err)
		}
		if !createIndex.Acknowledged {
			log.Println(createIndex)
		} else {
			log.Println("successfully created index")
		}
	} else {
		log.Println("Index already exist")
	}
}

func GetESClient() (*elastic.Client, error) {

	client, err :=  elastic.NewClient(elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))

	fmt.Println("ES initialized...")

	return client, err

}

func SaveLabels (labels ImageLabels)  {
	esclient, err := GetESClient()
	if err != nil {
		fmt.Println("Error initializing : ", err)
		panic("Client fail ")
	}

	dataJSON, errs := json.Marshal(labels)
	if errs != nil {
		panic(errs)
	}
	js := string(dataJSON)
	i, err := esclient.Index().
		Index(index).
		BodyJson(js).
		Do(ctx)

	if err != nil {
		panic(err)
	}
	fmt.Println(i)
	fmt.Println("[Elastic][InsertProduct]Insertion Successful")

}