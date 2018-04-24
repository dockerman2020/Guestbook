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

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const (
	storageProject = "STORAGE_PROJECT"
	storageBucket  = "STORAGE_BUCKET"
)

func main() {
	project := os.Getenv(storageProject)
	if project == "" {
		log.Fatalf("%s environment variable unspecified. Please provide Storage project.", storageProject)
	}

	bucketName := os.Getenv(storageBucket)
	if bucketName == "" {
		log.Fatalf("%s environment variable unspecified. Please provide Storage bucket.", storageBucket)
	}

	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	it := client.Bucket(bucketName).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading next object: %v", err)
			return
		}

		objectName := attrs.Name
		if err := client.Bucket(bucketName).Object(objectName).Delete(ctx); err != nil {
			log.Printf("Failed to delete object %q in bucket %q within project %q: %v", objectName, bucketName, project, err)
		} else {
			log.Printf("Succesfully deleted object %q in bucket %q within project %q", objectName, bucketName, project)
		}
	}
}
