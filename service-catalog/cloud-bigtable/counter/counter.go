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
	"bytes"
	"encoding/binary"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/bigtable"
	"golang.org/x/net/context"
)

var (
	client *bigtable.Client
	tbl    *bigtable.Table
)

const (
	port              = "8080"
	gcpProjectEnvName = "GOOGLE_CLOUD_PROJECT"
	btInstanceEnvName = "BIGTABLE_INSTANCE"
	tableName         = "visits"
	familyName        = "visits"
	columnName        = "visits"
)

var tmpl = template.Must(template.New("").Parse(`
<!doctype html><title>Bigtable example</title>
<h1>Bigtable example</h1>
<p><code>{{.Path}}</code> has been visited {{.Visits}} times.
`))

// mainHandler tracks how many times this page has been visited.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rmw := bigtable.NewReadModifyWrite()
	rowKey := r.URL.EscapedPath()
	rmw.Increment(familyName, columnName, 1)
	row, err := tbl.ApplyReadModifyWrite(ctx, rowKey, rmw)
	if err != nil {
		http.Error(w, fmt.Sprint("Error applying ReadModifyWrite to row:", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Path   string
		Visits uint64
	}{
		Path:   rowKey,
		Visits: binary.BigEndian.Uint64(row[familyName][0].Value),
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		http.Error(w, fmt.Sprint("Template error:", err), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func main() {
	projectID := os.Getenv(gcpProjectEnvName)
	if projectID == "" {
		log.Fatalf("Couldn't find %s in env", gcpProjectEnvName)
	}
	instance := os.Getenv(btInstanceEnvName)
	if instance == "" {
		log.Fatalf("Couldn't find %s in env", btInstanceEnvName)
	}

	ctx := context.Background()
	client, err := bigtable.NewClient(ctx, projectID, instance)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	tbl = client.Open(tableName)

	http.HandleFunc("/", mainHandler)
	log.Println("Listening on port:", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
