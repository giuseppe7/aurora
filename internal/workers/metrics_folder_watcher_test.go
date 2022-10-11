package workers

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var mfwWorker *MetricsFolderWatcher

const mfwPath = "../../test/data"

func TestMetricsFolderWatcherForUntyped(t *testing.T) {
	filename := "testCreateWriteRemoveUntyped.txt"

	if mfwWorker == nil {
		mfwWorker = NewMetricsFolderWatcher(mfwPath)
	}
	mfwWorker.WatchAndEmit()

	// Create
	testCreate := filepath.Join(mfwPath, filename)
	f, err := os.Create(testCreate)
	if err != nil {
		t.Errorf("expected to be able to create metric text file. %v", err)
		return
	}
	defer f.Close()

	// Write
	w := bufio.NewWriter(f)
	_, err = w.WriteString("test_untyped 123.45\n")
	if err != nil {
		t.Errorf("expected to be able to write metric text file. %v", err)
		return
	}
	w.Flush()
	time.Sleep(2 * time.Second)

	// Remove
	e := os.Remove(testCreate)
	if e != nil {
		t.Errorf("expected to be able to delete metric text file. %v", err)
		return
	}
}

func TestMetricsFolderWatcherForGauge(t *testing.T) {
	filename := "testCreateWriteRemoveGauge.txt"
	path := "../../test/data"

	if mfwWorker == nil {
		mfwWorker = NewMetricsFolderWatcher(mfwPath)
	}
	mfwWorker.WatchAndEmit()

	// Create
	testCreate := filepath.Join(path, filename)
	f, err := os.Create(testCreate)
	if err != nil {
		t.Errorf("expected to be able to create metric text file. %v", err)
		return
	}
	defer f.Close()

	// Write
	w := bufio.NewWriter(f)
	text := `
	# TYPE test_gauge gauge
	test_gauge 123.45
`
	_, err = w.WriteString(text)
	if err != nil {
		t.Errorf("expected to be able to write metric text file. %v", err)
		return
	}
	w.Flush()
	time.Sleep(2 * time.Second)

	// Remove
	e := os.Remove(testCreate)
	if e != nil {
		t.Errorf("expected to be able to delete metric text file. %v", err)
		return
	}
}
