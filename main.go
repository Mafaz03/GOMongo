package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CatFactWorkerServer struct {
	client *mongo.Client
}

func NewCatWorkerServer(c *mongo.Client) *CatFactWorkerServer {
	return &CatFactWorkerServer{
		client: c,
	}
}

func (s *CatFactWorkerServer) getFacts(w http.ResponseWriter, r *http.Request) {
	coll := s.client.Database("CatFacts").Collection("Facts")
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	results := []bson.M{}
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)

}


type CatFactWorker struct {
	client *mongo.Client
}

func NewCatWorker(c *mongo.Client) *CatFactWorker {
	return &CatFactWorker{
		client: c,
	}
}

func (cfw *CatFactWorker) start() error {
	coll := cfw.client.Database("CatFacts").Collection("Facts")
	ticker := time.NewTicker(2 * time.Second)
	for {
		response, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}
		var catFact bson.M
		err = json.NewDecoder(response.Body).Decode(&catFact)
		if err != nil {
			log.Fatal(err)
		}
		_, err = coll.InsertOne(context.TODO(), catFact)
		if err != nil {
			return err
		}
		fmt.Println(catFact["fact"])
		<-ticker.C
	}
}

func main() {
	mogno_client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Println("Error: ", err)
	}
	worker := NewCatWorker(mogno_client)
	go worker.start()
	
	server := NewCatWorkerServer(mogno_client)
	http.HandleFunc("/facts", server.getFacts)
	http.ListenAndServe(":8080", nil)
}
