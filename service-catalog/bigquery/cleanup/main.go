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
	"log"
	"os"
	"time"
	"net/http"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
)

const (
	bigQueryProject = "BIGQUERY_PROJECT"
	bigQueryDataset = "BIGQUERY_DATASET"

	tableID = "github_contributors"
)

func main() {
	project := os.Getenv(bigQueryProject)
	if project == "" {
		log.Fatalf("%s environment variable unspecified. Please provide BigQuery project.", bigQueryProject)
	}

	datasetID := os.Getenv(bigQueryDataset)
	if datasetID == "" {
		log.Fatalf("%s environment variable unspecified. Please provide BigQuery dataset ID.", bigQueryDataset)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client, err := bigquery.NewClient(ctx, project)
	if err != nil {
		log.Fatalf("Failed to create bigquery client: %v", err)
	}

	if err := client.Dataset(datasetID).Table(tableID).Delete(ctx); err != nil {
		gapiErr, ok := err.(*googleapi.Error)
		if ok {
			if gapiErr.Code == http.StatusNotFound {
				log.Println("Nothing to delete.")
				return
			}
		}
		log.Fatalf("Failed to delete user table: %v", err)
	}
	log.Println("Successfully deleted user table.")
}
