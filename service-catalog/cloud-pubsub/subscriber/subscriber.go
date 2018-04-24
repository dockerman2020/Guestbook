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
	"html"
	"log"
	"net/http"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
)

const (
	port              = "8080"
	gcpProjectEnvName = "GOOGLE_CLOUD_PROJECT"
	pubsubSubEnvName  = "PUBSUB_SUBSCRIPTION"
)

var (
	messages     []*pubsub.Message
	messagesLock sync.RWMutex
)

func handleListMessages(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<!DOCTYPE html><title>Pubsub example</title>",
		"<h1>Pubsub example</h1>",
		"<p>Received messages:</p>",
		"<ul>")
	messagesLock.RLock()
	defer messagesLock.RUnlock()
	for _, m := range messages {
		fmt.Fprintln(w, "<li>", html.EscapeString(string(m.Data)))
	}
}

func main() {
	cctx, cancel := context.WithCancel(context.Background())
	go receiveMessages(cctx)
	defer cancel()

	http.HandleFunc("/", handleListMessages)
	log.Println("Listening on port: ", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func receiveMessages(ctx context.Context) {
	projectID := os.Getenv(gcpProjectEnvName)
	if projectID == "" {
		log.Fatalf("Couldn't find %s in env", gcpProjectEnvName)
	}

	subName := os.Getenv(pubsubSubEnvName)
	if subName == "" {
		log.Fatalf("Couldn't find %s in env", pubsubSubEnvName)
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	sub := client.Subscription(subName)

	ok, err := sub.Exists(ctx)
	if err != nil {
		log.Fatalf("Error finding subscription %s: %v", subName, err)
	}
	if !ok {
		log.Fatalf("Couldn't find subscription %v", subName)
	}

	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		log.Printf("Got message ID=%s, payload=[%s]", m.ID, m.Data)
		messagesLock.Lock()
		messages = append(messages, m)
		messagesLock.Unlock()
		m.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to receive: %v", err)
	}
}
