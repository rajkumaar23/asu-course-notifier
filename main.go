package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"errors"
	"github.com/rollbar/rollbar-go"
)

var WaitGroup sync.WaitGroup
var CourseCatalogBaseURL = "https://eadvs-cscc-catalog-api.apps.asu.edu/catalog-microservices/api/v1/search/classes"
var HttpClient = &http.Client{}
var config Config
var rollbarToken string

func init() {
	rollbarToken = os.Getenv("ROLLBAR_TOKEN")
	if rollbarToken != "" {
		rollbar.SetToken(rollbarToken)
		rollbar.SetEnvironment(os.Getenv("ROLLBAR_ENV"))
	}
}

func main() {
	defer func() {
		rollbar.Close()
	}()
	configFileContents, err := os.ReadFile("config.json")
	if err != nil {
		LogErrorAndPanic(errors.Join(errors.New("Error reading config file contents"), err))
	}
	if err = json.Unmarshal(configFileContents, &config); err != nil {
		LogErrorAndPanic(errors.Join(errors.New("Error parsing config file as JSON"), err))
	}

	for class := range config.CoursesToWatch {
		fmt.Println("Checking availability for", class)
		WaitGroup.Add(1)
		go CheckAvailability(class)
	}

	WaitGroup.Wait()
}
