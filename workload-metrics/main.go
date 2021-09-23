// # Copyright 2021 Google LLC
// #
// # Licensed under the Apache License, Version 2.0 (the "License");
// # you may not use this file except in compliance with the License.
// # You may obtain a copy of the License at
// #
// #     http://www.apache.org/licenses/LICENSE-2.0
// #
// # Unless required by applicable law or agreed to in writing, software
// # distributed under the License is distributed on an "AS IS" BASIS,
// # WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// # See the License for the specific language governing permissions and
// # limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	logger, _    = zap.NewProduction()
	reg          = prometheus.NewRegistry()
	requestCount = promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "example_requests_total",
			Help: "Total number of HTTP requests by status code and method.",
		},
		[]string{"code", "method"},
	)
	histogram = promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
		Name: "example_random_numbers",
		Help: "A histogram of normally distributed random numbers.",
		// TODO: support negative bounds
		// Buckets: prometheus.LinearBuckets(-3, .1, 61),
		Buckets: prometheus.LinearBuckets(0, .1, 61),
	})
)

// Generates random data for a histogram
func Random() {
	logger.Sugar().Info("Started number generator")
	for {
		val := rand.NormFloat64()
		if val < 0 {
			val = -val
		}
		histogram.Observe(val)
	}
}

// PollItself polls the HTTP endpoint to generate synthetic "traffic"
func PollItself() {
	for {
		resp, err := http.Get("http://localhost:1234/")
		if err != nil {
			logger.Sugar().Errorf("HTTP request failed: %w", err)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Sugar().Errorf("Failed to read response: %w", err)
			} else {
				logger.Sugar().Debugf("Response: %s", string(body))
				_ = resp.Body.Close()
			}
		}
		time.Sleep(time.Second * time.Duration(rand.Intn(3)))
	}
}

func parseFlags() {
	enableProcessMetrics := flag.Bool("process-metrics", false, "Enables process metrics")
	enableGoMetrics := flag.Bool("go-metrics", false, "Enables Go metrics")

	flag.Parse()

	if *enableProcessMetrics {
		reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}

	if *enableGoMetrics {
		reg.MustRegister(prometheus.NewGoCollector())
	}
}

func main() {
	parseFlags()
	go Random()
	go PollItself()

	// Example HTTP handler
	http.Handle("/", promhttp.InstrumentHandlerCounter(
		requestCount,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			random := rand.Intn(100)
			if random < 20 {
				w.WriteHeader(500)
				_, _ = fmt.Fprint(w, "Something went wrong :(")
				return
			}
			_, _ = fmt.Fprint(w, "Hello, world!")
		}),
	))
	// Expose Prometheus metrics
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	logger.Info("Starting HTTP server")
	err := http.ListenAndServe(":1234", nil)

	if err != nil {
		logger.Sugar().Errorf("ListenAndServe failed: %w", err)
	}
}
