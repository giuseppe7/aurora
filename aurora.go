package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/giuseppe7/aurora/internal/workers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const applicationNamespace = "aurora"
const defaultWatchedFolder = "./test/data"

// Variable to be set by the Go linker at build time.
var version string

// Set up observability with Prometheus handler for metrics.
func initObservability() {

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	// Register a version gauge.
	versionGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: applicationNamespace,
			Name:      "version_info",
			Help:      "Version of the application.",
		},
	)
	prometheus.MustRegister(versionGauge)
	versionValue, err := strconv.ParseFloat(version, 64)
	if err != nil {
		versionValue = 0.0
	}
	versionGauge.Set(versionValue)
}

func main() {
	log.Println("Coming online...")
	log.Println("Version:", version)

	// Channel to be aware of an OS interrupt like Control-C.
	var waiter sync.WaitGroup
	waiter.Add(1)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Set up observability.
	initObservability()

	// Set up file watcher to detect and create metrics.
	path := defaultWatchedFolder
	if os.Getenv("WATCHED_DIR") != "" {
		path = os.Getenv(("WATCHED_DIR"))
	}

	// Set up file watcher to detect and create metrics.
	log.Println("Folder being watched:", path)
	w := workers.NewMetricsFolderWatcher(path) // TODO: Singleton?
	w.WatchAndEmit()

	// Function and waiter to wait for the OS interrupt and do any clean-up.
	log.Println("Running.")
	go func() {
		<-c
		fmt.Println("\r")
		log.Println("Interrupt captured.")
		waiter.Done()
	}()
	waiter.Wait()

	// Shut down the application.
	log.Println("Shutting down.")
}
