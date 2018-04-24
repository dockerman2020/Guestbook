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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const (
	storageBucket = "STORAGE_BUCKET"
	port          = "8080"
)

var bucketName string

type quote struct {
	Person string `json:"person,omitempty"`
	Quote  string `json:"quote,omitempty"`
}

type quotes struct {
	Quotes []quote `json:"quotes"`
}

func main() {
	bucketName = os.Getenv(storageBucket)
	if bucketName == "" {
		log.Fatalf("%s environment variable unspecified. Please provide Storage bucket.", storageBucket)
	}

	r := mux.NewRouter()
	r.HandleFunc("/quotes", handleCreateQuote).Methods("POST")
	r.HandleFunc("/quotes", handleListQuotes).Methods("GET")

	log.Println("Server listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleCreateQuote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		writeError(w, http.StatusBadRequest, "cannot read the request body")
		return
	}
	var e quote
	if err := json.Unmarshal(body, &e); err != nil {
		log.Printf("Error parsing request body: %v", err)
		writeError(w, http.StatusBadRequest, "cannot parse the request body - invalid JSON")
		return
	}
	if e.Person == "" {
		writeError(w, http.StatusBadRequest, "missing required property: \"Person\"")
		return
	}
	if e.Quote == "" {
		writeError(w, http.StatusBadRequest, "missing required property: \"Quote\"")
		return
	}

	wc := client.Bucket(bucketName).Object(e.Person).NewWriter(ctx)
	if _, err := wc.Write([]byte(e.Quote)); err != nil {
		log.Printf("Unable to write quote %q to bucket %q, object %q: %v", e.Quote, bucketName, e.Person, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := wc.Close(); err != nil {
		log.Printf("Unable to close writer: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func handleListQuotes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	it := client.Bucket(bucketName).Objects(ctx, nil)
	e := quotes{Quotes: []quote{}}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading next object: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		objectName := attrs.Name
		q, err := readObject(ctx, client, objectName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		e.Quotes = append(e.Quotes, quote{Person: objectName, Quote: q})
	}

	j, err := json.Marshal(e)
	if err != nil {
		log.Printf("Error marshaling JSON for input %+v: %v", e, err)
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

func readObject(ctx context.Context, client *storage.Client, objectName string) (string, error) {
	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		log.Printf("Error creating a reader for bucket %q, object %q: %v", bucketName, objectName, err)
		return "", err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Printf("Error reading contents of bucket %q, object %q: %v", objectName, bucketName, err)
		return "", err
	}
	return string(data), nil
}
