package workers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const metricsFileExtension = ".txt"

type MetricsFolderWatcher struct {
	watcher        *fsnotify.Watcher
	path           string
	eventsGaugeVec *prometheus.GaugeVec
	adhocGauges    map[string]prometheus.Gauge
}

func NewMetricsFolderWatcher(path string) *MetricsFolderWatcher {
	worker := new(MetricsFolderWatcher)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to create NewWatcher: ", err)
	}

	labels := []string{"filename", "operations"}
	eventsGaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "metrics_folder_watcher_events",
			Help: "Events detected by the Metrics Folder Watcher.",
		},
		labels,
	)
	prometheus.MustRegister(eventsGaugeVec)

	worker.watcher = watcher
	worker.path = path
	worker.eventsGaugeVec = eventsGaugeVec
	worker.adhocGauges = make(map[string]prometheus.Gauge)
	return worker
}

func (worker *MetricsFolderWatcher) WatchAndEmit() {
	done := make(chan bool)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-worker.watcher.Events:
				if !ok {
					return
				}
				worker.respondToEvent(event)
			case err, ok := <-worker.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}

	}()
	err := worker.watcher.Add(worker.path)
	if err != nil {
		log.Fatal("Failed to add a folder to the watcher:", err)
	}
}

func (worker *MetricsFolderWatcher) respondToEvent(event fsnotify.Event) {
	basename := filepath.Base(event.Name)
	worker.eventsGaugeVec.WithLabelValues(basename, event.Op.String()).Inc()

	// Ignore everything but .txt files.
	var extension = filepath.Ext(event.Name)
	if extension != metricsFileExtension {
		return
	}

	if event.Op == fsnotify.Chmod {
		// Ignore chmod for now.
	} else if event.Op == fsnotify.Rename {
		// Ignore rename for now.
	} else if event.Op == fsnotify.Create {
		log.Printf("Create: %s %s\n", event.Name, event.Op)
		worker.parseAndUpsert(event)
	} else if event.Op == fsnotify.Write {
		log.Printf("Write: %s %s\n", event.Name, event.Op)
		worker.parseAndUpsert(event)
	} else if event.Op == fsnotify.Remove {
		// Ignore remove for now.
	} else {
		log.Printf("Unknown: %s %s\n", event.Name, event.Op)
	}
}

func (worker *MetricsFolderWatcher) parseAndUpsert(event fsnotify.Event) {
	basename := filepath.Base(event.Name)
	data, err := os.ReadFile(event.Name)
	if err != nil {
		log.Println("Error in reading file", basename, err)
		return
	}

	var parser expfmt.TextParser
	parsed, err := parser.TextToMetricFamilies(strings.NewReader(string(data)))
	if err != nil {
		log.Println("Error in parsing file", basename, err)
		return
	}

	// Iterate throught the metrics family found...
	for _, mf := range parsed {
		// TODO: Everything is a gauge for now.
		if _, ok := worker.adhocGauges[mf.GetName()]; !ok {
			// If not known, track and register.
			adhocGauge := prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: mf.GetName(),
					Help: mf.GetHelp(),
				},
			)
			prometheus.MustRegister(adhocGauge)
			worker.adhocGauges[mf.GetName()] = adhocGauge
		}

		for _, metric := range mf.GetMetric() {
			if _, ok := worker.adhocGauges[mf.GetName()]; ok {
				value := 0.0
				if mf.GetType() == io_prometheus_client.MetricType_GAUGE {
					value = *metric.Gauge.Value
				} else if mf.GetType() == io_prometheus_client.MetricType_COUNTER {
					value = *metric.Counter.Value
				} else if mf.GetType() == io_prometheus_client.MetricType_UNTYPED {
					value = *metric.Untyped.Value
				}
				log.Print(fmt.Sprintf(" Faking gauge %s, %s, with %f", mf.GetName(), mf.GetType(), value))
				worker.adhocGauges[mf.GetName()].Set(value)
			}
		}

		/*
			log.Println(mf)
			log.Print(fmt.Sprintf(" - name: %s", mf.GetName()))
			log.Print(fmt.Sprintf(" - type: %s", mf.GetType()))
			for _, metric := range mf.GetMetric() {
				for _, label := range metric.GetLabel() {
					log.Print(fmt.Sprintf(" - label: %s", label))
				}
				if mf.GetType() == io_prometheus_client.MetricType_GAUGE {
					log.Print(fmt.Sprintf(" - value: %f", metric.GetGauge().GetValue()))
				} else {
					log.Print(fmt.Sprintf(" - value: %f", metric.GetUntyped().GetValue()))
				}
			}
		*/
	}
}
