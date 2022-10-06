package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
	"time"
)

func TestObservability(t *testing.T) {

	initObservability()

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 10 * time.Second,
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	apiEndpoint := "/metrics"
	url := fmt.Sprintf("http://0.0.0.0:2112/%s", apiEndpoint)

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Errorf("expected to be able to reach the metrics endpoint")
		return
	} else if resp.StatusCode != 200 {
		t.Errorf("connected but received non-200 status code %d", resp.StatusCode)
		return
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unexpected error reading the response body")
		return
	}

	bodyString := string(bodyBytes)
	r := regexp.MustCompile(`.*_version_info (\d.*?)`)
	if len(r.FindStringSubmatch(bodyString)) == 0 {
		t.Errorf("expecting to find the version_info in the metrics endpoint")
	}
	defer resp.Body.Close()
}
