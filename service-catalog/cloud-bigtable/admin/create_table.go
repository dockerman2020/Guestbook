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

	"cloud.google.com/go/bigtable"
	"golang.org/x/net/context"
)

const (
	gcpProjectEnvName = "GOOGLE_CLOUD_PROJECT"
	btInstanceEnvName = "BIGTABLE_INSTANCE"
	tableName         = "visits"
	familyName        = "visits"
)

func main() {
	projectID := os.Getenv(gcpProjectEnvName)
	if projectID == "" {
		log.Fatalf("Couldn't find %s in env", gcpProjectEnvName)
	}
	instance := os.Getenv(btInstanceEnvName)
	if instance == "" {
		log.Fatalf("Couldn't find %s  in env", btInstanceEnvName)
	}
	err := createTable(context.Background(), projectID, instance)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	log.Println("Successfully created table", tableName)
}

// createTable sets up admin client, tables, and column families.
func createTable(ctx context.Context, project string, instance string) error {
	// NewAdminClient uses Application Default Credentials to authenticate.
	adminClient, err := bigtable.NewAdminClient(ctx, project, instance)
	if err != nil {
		return fmt.Errorf("Unable to create a table admin client: %v", err)
	}
	defer adminClient.Close()
	tables, err := adminClient.Tables(ctx)
	if err != nil {
		return fmt.Errorf("Unable to fetch table list: %v", err)
	}
	if !sliceContains(tables, tableName) {
		if err := adminClient.CreateTable(ctx, tableName); err != nil {
			return fmt.Errorf("Unable to create table: %v. %v", tableName, err)
		}
	}
	tblInfo, err := adminClient.TableInfo(ctx, tableName)
	if err != nil {
		return fmt.Errorf("Unable to read info for table: %v. %v", tableName, err)
	}
	if !sliceContains(tblInfo.Families, familyName) {
		if err := adminClient.CreateColumnFamily(ctx, tableName, familyName); err != nil {
			return fmt.Errorf("Unable to create column family: %v. %v", familyName, err)
		}
	}
	return nil
}

// sliceContains reports whether the provided string is present in the given slice of strings.
func sliceContains(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}
