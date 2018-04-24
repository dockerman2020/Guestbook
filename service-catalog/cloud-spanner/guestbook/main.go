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
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"

	"cloud.google.com/go/spanner"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const (
	spannerProject  = "SPANNER_PROJECT"
	spannerInstance = "SPANNER_INSTANCE"
	databaseName    = "guestbook"
	port            = "8080"
)

var (
	dbClient *spanner.Client
)

type entry struct {
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
}

type entries struct {
	Entries []entry `json:"entries"`
}

func main() {
	project := os.Getenv(spannerProject)
	if project == "" {
		log.Fatal(spannerProject + " environment variable unspecified. Please provide Spanner project.")
	}

	instance := os.Getenv(spannerInstance)
	if instance == "" {
		log.Fatal(spannerInstance + " environment variable unspecified. Please provide Spanner instance.")
	}

	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, databaseName)
	log.Printf("Using database %q", db)

	var err error
	dbClient, err = spanner.NewClient(context.Background(), db)
	if err != nil {
		log.Fatalf("Failed to create a database client: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/entries", handleListEntries).Methods("GET")
	r.HandleFunc("/entries", handleCreateEntry).Methods("POST")

	log.Println("Server listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleListEntries(w http.ResponseWriter, r *http.Request) {
	iter := dbClient.Single().Read(context.TODO(), "Guestbook", spanner.AllKeys(), []string{"Name", "Message"})
	defer iter.Stop()

	e := entries{Entries: []entry{}}
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Printf("Error reading net row: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		var name, message string
		if err := row.Columns(&name, &message); err != nil {
			log.Printf("Error reading column: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		e.Entries = append(e.Entries, entry{Name: name, Message: message})
	}

	j, err := json.Marshal(e)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		writeError(w, http.StatusBadRequest, "cannot read the request body")
		return
	}
	var e entry
	if err := json.Unmarshal(body, &e); err != nil {
		log.Printf("Error parsing request body: %v", err)
		writeError(w, http.StatusBadRequest, "cannot parse the reqeust body - invalid JSON")
		return
	}

	if e.Name == "" {
		writeError(w, http.StatusBadRequest, "missing required property: \"name\"")
		return
	}

	if e.Message == "" {
		writeError(w, http.StatusBadRequest, "missing required property: \"message\"")
		return
	}

	id, err := rand.Int(rand.Reader, (&big.Int{}).SetInt64(math.MaxInt64))
	if err != nil {
		log.Printf("Error generating random ID: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	cols := []string{"Id", "Name", "Message"}
	if _, err := dbClient.Apply(context.TODO(), []*spanner.Mutation{
		spanner.Insert("Guestbook", cols, []interface{}{id.Int64(), e.Name, e.Message}),
	}); err != nil {
		log.Printf("Error writing into Spanner: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
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
