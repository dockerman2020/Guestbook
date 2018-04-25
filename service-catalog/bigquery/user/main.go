/**
 * Copyright 2018 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const (
	bigQueryProject = "BIGQUERY_PROJECT"
	bigQueryDataset = "BIGQUERY_DATASET"

	tableID = "github_contributors"

	port = "8080"
)

var (
	client    *bigquery.Client
	datasetID string
)

type entry struct {
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
}

type entries struct {
	Entries []entry `json:"entries"`
}

func main() {
	project := os.Getenv(bigQueryProject)
	if project == "" {
		log.Fatalf("%s environment variable unspecified. Please provide BigQuery project.", bigQueryProject)
	}

	datasetID = os.Getenv(bigQueryDataset)
	if datasetID == "" {
		log.Fatalf("%s environment variable unspecified. Please provide BigQuery dataset ID.", bigQueryDataset)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var err error
	client, err = bigquery.NewClient(ctx, project)
	if err != nil {
		log.Fatalf("Failed to create bigquery client: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/query", handleQuery).Methods("GET")

	log.Println("Server listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	prefix := "this actually works"
	if p, ok := r.URL.Query()["prefix"]; ok {
		if len(p) >= 1 {
			// Just use the first `?prefix=value` value.
			prefix = p[0]
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	tableLoc := datasetID + "." + tableID
	q := client.Query(
		"SELECT committer.name as name, message " +
				"FROM `" + tableLoc + "` " +
				"WHERE STARTS_WITH(message, '" + prefix + "') " +
				"GROUP BY name, message")
	it, err := q.Read(ctx)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var res entries
	for {
		var e entry
		err := it.Next(&e)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading next row: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		res.Entries = append(res.Entries, e)
	}

	j, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	j, err := json.Marshal(
		struct {
			Error string `json:"error"`
		}{
			Error: message,
		})
	if err != nil {
		w.Write([]byte("{\"error\": \"internal server error\"}"))
		return
	}
	w.Write(j)
}
