package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
)

var WaitGroup sync.WaitGroup
var CourseCatalogBaseURL = "https://eadvs-cscc-catalog-api.apps.asu.edu/catalog-microservices/api/v1/search/classes"
var HttpClient = &http.Client{}
var config Config

func main() {
	configFileContents, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(configFileContents, &config); err != nil {
		panic(err)
	}

	for class := range config.CoursesToWatch {
		PrintlnWithPrefixedTime("Checking availability for", class)
		WaitGroup.Add(1)
		go CheckAvailability(class)
	}

	WaitGroup.Wait()
}
