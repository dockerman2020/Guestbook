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
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
)

const (
	port               = "8080"
	gcpProjectEnvName  = "GOOGLE_CLOUD_PROJECT"
	pubsubTopicEnvName = "PUBSUB_TOPIC"
)

var topic *pubsub.Topic

func main() {
	ctx := context.Background()

	projectID := os.Getenv(gcpProjectEnvName)
	if projectID == "" {
		log.Fatalf("Couldn't find %s in env", gcpProjectEnvName)
	}

	topicName := os.Getenv(pubsubTopicEnvName)
	if topicName == "" {
		log.Fatalf("Couldn't find %s in env", pubsubTopicEnvName)
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	log.Println("Created client")

	topic = client.Topic(topicName)

	// The topic existence test requires the binding to have the 'viewer' role.
	ok, err := topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Error finding topic: %v", err)
	}
	if !ok {
		log.Fatalf("Couldn't find topic %v", topic)
	}
	defer topic.Stop()

	http.HandleFunc("/", getHandler)
	http.HandleFunc("/publish", postHandler)

	log.Println("Listening on port:", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	message := r.FormValue("message")
	result := topic.Publish(ctx, &pubsub.Message{Data: []byte(message)})
	serverID, err := result.Get(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Published message ID=%s", serverID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<!doctype html><form method='POST' action='/publish'>"+
		"<input required name='message' placeholder='Message'>"+
		"<input type='submit' value='Publish'>"+
		"</form>")
}
