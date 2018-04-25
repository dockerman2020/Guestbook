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

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
)

const (
	bigQueryProject = "BIGQUERY_PROJECT"
	bigQueryDataset = "BIGQUERY_DATASET"

	tableID = "github_contributors"

	publicDataProject  = "bigquery-public-data"
	githubReposDataset = "github_repos"
	commitsTable       = "commits"
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

	dstTable := client.Dataset(datasetID).Table(tableID)
	srcTable := client.DatasetInProject(publicDataProject, githubReposDataset).Table(commitsTable)

	copier := dstTable.CopierFrom(srcTable)
	// WriteTruncate overrides the existing data in the destination table.
	// Data is overwritten atomically on successful completion of a job.
	copier.WriteDisposition = bigquery.WriteTruncate

	// This will create the destination table if it does not already exist.
	job, err := copier.Run(ctx)
	if err != nil {
		log.Fatalf("Failed to start copy job: %v", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch status of copy job: %v", err)
	}
	if status.Err() != nil {
		log.Fatalf("Copy job failed: %v", err)
	}
	log.Println("Data successfully copied to user table.")
}
