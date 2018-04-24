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
	"os"
	"time"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"golang.org/x/net/context"
	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	spannerProject  = "SPANNER_PROJECT"
	spannerInstance = "SPANNER_INSTANCE"
	databaseName    = "guestbook"
)

func createDatabase(ctx context.Context, ac *database.DatabaseAdminClient, project, instance string) error {
	op, err := ac.CreateDatabase(ctx, &adminpb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", project, instance),
		CreateStatement: "CREATE DATABASE `" + databaseName + "`",
		ExtraStatements: []string{
			`CREATE TABLE Guestbook (
				Id INT64 NOT NULL,
				Name STRING(64) NOT NULL,
				Message STRING(MAX) NOT NULL,
			) PRIMARY KEY (Id)`,
		},
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			// Database already exists
			return nil
		}
		log.Printf("Failed the CreateDatabase call: %v", err)
		return err
	}
	if _, err := op.Wait(ctx); err != nil {
		log.Printf("Failed the Operation Wait call: %v", err)
		return err
	}
	return nil
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	log.Printf("Creating database in %q project, %q instance", project, instance)

	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create database admin client: %v", err)
	}

	if err := createDatabase(ctx, adminClient, project, instance); err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	log.Println("Database created successfully.")
}
