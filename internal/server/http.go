package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"apatit/internal/cache"
)

// startServer runs HTTP-server.
func StartServer(listenAddress string) {
	// JSON stats endpoint
	http.HandleFunc("/stats", statsHandler)

	// Metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		_, _ = w.Write([]byte(`
<html><head><title>APATIT</title></head><body>
<h1>APATIT</h1>
<h2>Advanced Ping-Admin Task Indicators Transducer</h2>
<p><a href='/metrics'>Metrics</a></p>
<p><a href='/stats?type=task'>Tasks JSON</a></p>
<p><a href='/stats?type=all'>All Tasks Info JSON</a></p>
</body></html>`))
	})

	logrus.WithField("address", listenAddress).Info("Starting HTTP server")
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		logrus.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// statsHandler handle /stats request with 'type' parameter.
func statsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	dataType := queryParams.Get("type")

	var jsonData []byte

	switch dataType {
	case "task":
		jsonData = cache.TaskDataCache.GetFromCache()
	case "all":
		jsonData = cache.AllTasksInfoCache
	default:
		jsonData = []byte(`{"error":"Invalid or missing 'type' parameter. Use 'type=task' or 'type=all'."}`)
	}

	if len(jsonData) == 0 {
		jsonData = []byte("[]")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write(jsonData)
	if err != nil {
		logrus.Errorf("Failed to write response: %v", err)
	}
}
