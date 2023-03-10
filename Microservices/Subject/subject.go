package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// http://localhost:5053/api/getsubjects/all
// {"A-level": ["a","b"], "O-level": [...]..}
// http://localhost:5053/api/getsubjects/psle
// ["a","b",..]
// http://localhost:5053/api/getsubjects/olevel
// http://localhost:5053/api/getsubjects/alevel

var cred_file = "/eti-assignment-2-firebase-adminsdk-6r9lk-85fb98eda4.json"

// var url = "https://react-app-4dcnj7fm6a-uc.a.run.app
//var url = "http://localhost:3000"
var url = "http://104.154.110.27"

func Subject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", url)

	params := mux.Vars(r)
	req_type := params["type"]

	ctx := context.Background()

	// Use a service account
	sa := option.WithCredentialsFile(cred_file)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	// get all the data
	iter := client.Collection("Global Data").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		data := doc.Data()
		if data != nil {
			if req_type == "all" {
				json.NewEncoder(w).Encode(data)
				return
			} else if req_type == "psle" {
				for k := range data {
					if k == "PSLE" {
						json.NewEncoder(w).Encode(data[k])
					}
				}
				return
			} else if req_type == "olevel" {
				for k := range data {
					if k == "O-Level" {
						json.NewEncoder(w).Encode(data[k])
					}
				}
				return
			} else if req_type == "alevel" {
				for k := range data {
					if k == "A-Level" {
						json.NewEncoder(w).Encode(data[k])
					}
				}
				return
			} else {
				w.WriteHeader(http.StatusNotAcceptable) // 406
				return
			}
		}
	}
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/getsubjects/{type}", Subject).Methods("GET", "OPTIONS")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{url},
		AllowCredentials: true,
	})

	handler := cors.Default().Handler(router)
	handler = c.Handler(handler)

	fmt.Println("Listening at port 5051")
	log.Fatal(http.ListenAndServe(":5051", handler))
}
