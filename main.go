package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

type Class struct {
	Details struct {
		ClassNumber string   `json:"CLASSNBR"`
		Title       string   `json:"TITLE"`
		Term        string   `json:"STRM"`
		Instructors []string `json:"INSTRUCTORSLIST"`
	} `json:"CLAS"`
	SeatInfo struct {
		EnrollmentCap   int `json:"ENRL_CAP"`
		EnrollmentTotal int `json:"ENRL_TOT"`
	} `json:"seatInfo"`
	ReservedSeatsInfo []struct {
		EnrollmentCap   int `json:"ENRL_CAP"`
		EnrollmentTotal int `json:"ENRL_TOT"`
	} `json:"reservedSeatsInfo"`
	SubjectNumber string `json:"SUBJECTNUMBER"`
}

type CourseCatalogResponse struct {
	Classes  []Class `json:"classes"`
	ScrollID string  `json:"scrollId"`
}

type Config struct {
	TelegramIDs    map[string]int      `json:"telegram_ids"`
	CoursesToWatch map[string][]string `json:"courses"`
	TermID         string              `json:"TERM_ID"`
	DepartmentCode string              `json:"DEPT_CODE"`
	BotToken       string              `json:"BOT_TOKEN"`
}

func GetParamsForCourseCatalog(config Config, scrollId string) string {
	params := url.Values{}
	params.Add("refine", "Y")
	params.Add("subject", config.DepartmentCode)
	params.Add("term", config.TermID)
	if scrollId != "" {
		params.Add("scrollId", scrollId)
	}
	return params.Encode()
}

func GetFormattedCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func NotifyUser(telegramID int, user string, message string, class Class, client *http.Client, config Config) {
	defer waitGroup.Done()

	params := url.Values{}
	params.Set("chat_id", strconv.Itoa(telegramID))
	params.Set("text", message)
	params.Set("parse_mode", "HTML")
	telegramURL := fmt.Sprintf("https://api.telegram.org/%v/sendMessage?%v", config.BotToken, params.Encode())

	req, _ := http.NewRequest(http.MethodGet, telegramURL, nil)
	res, err := client.Do(req)
	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()

	errorMessagePrefix := fmt.Sprintf("Sending message to %v for %v failed :", user, class.Details.Title)
	if res.StatusCode != 200 {
		fmt.Println(GetFormattedCurrentTime(), ":", errorMessagePrefix, string(body))
		return
	}
	if err != nil {
		fmt.Println(GetFormattedCurrentTime(), ":", errorMessagePrefix, err)
		return
	}

	fmt.Println(GetFormattedCurrentTime(), ":", "Sent message to", user, "for", class.Details.Title, "(", class.SubjectNumber, ")")
}

func ProcessClasses(result CourseCatalogResponse, config Config, client *http.Client) {
	for _, class := range result.Classes {
		reservedAvailability := 0
		var difference int
		for _, item := range class.ReservedSeatsInfo {
			difference = item.EnrollmentCap - item.EnrollmentTotal
			// To handle cases where ENRL_TOT > ENRL_CAP for some weird reason
			reservedAvailability += int(math.Max(float64(difference), 0))
		}

		availableSlots := class.SeatInfo.EnrollmentCap - class.SeatInfo.EnrollmentTotal - reservedAvailability
		if users, exists := config.CoursesToWatch[class.Details.ClassNumber]; exists && availableSlots > 0 {
			applyLink := fmt.Sprintf(
				"https://go.oasis.asu.edu/addclass/?STRM=%v&ASU_CLASS_NBR=%v",
				class.Details.Term,
				class.Details.ClassNumber,
			)
			swapLink := fmt.Sprintf("https://webapp4.asu.edu/myasu/?action=swapclass&strm=%v", config.TermID)
			for _, user := range users {
				telegramID, userExists := config.TelegramIDs[user]
				if !userExists {
					fmt.Println(GetFormattedCurrentTime(), ":", "Telegram ID for", user, "could not be found")
					continue
				}
				waitGroup.Add(1)
				message := fmt.Sprintf(
					"Hey %v,\n\n%v (%v) by %v is available with %v slots.\n\nClick %v to add to cart.\n\nClick %v to swap with another course.",
					user,
					class.Details.Title,
					class.SubjectNumber,
					class.Details.Instructors[0],
					availableSlots,
					applyLink,
					swapLink,
				)
				fmt.Println(GetFormattedCurrentTime(), ":", "Sending message to", user, "for", class.Details.Title, "(", class.SubjectNumber, ")")
				go NotifyUser(telegramID, user, message, class, client, config)
			}
		}
	}
}

var waitGroup sync.WaitGroup

func main() {
	configFileContents, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var config Config
	if err = json.Unmarshal(configFileContents, &config); err != nil {
		panic(err)
	}

	courseCatalogBaseURL := "https://eadvs-cscc-catalog-api.apps.asu.edu/catalog-microservices/api/v1/search/classes"
	courseCatalogURL := fmt.Sprintf(
		"%v?%v",
		courseCatalogBaseURL,
		GetParamsForCourseCatalog(config, ""),
	)

	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodGet, courseCatalogURL, nil)
	req.Header.Set("Authorization", "Bearer null")

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var result CourseCatalogResponse
	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	// Performing our iteration on the 'first' set of classes
	scrollId := result.ScrollID
	ProcessClasses(result, config, client)

	// Paginating based on 'scrollId' until we receive no more results
	for {
		courseCatalogURL = fmt.Sprintf(
			"%v?%v",
			courseCatalogBaseURL,
			GetParamsForCourseCatalog(config, scrollId),
		)

		req, _ = http.NewRequest(http.MethodGet, courseCatalogURL, nil)
		req.Header.Set("Authorization", "Bearer null")

		res, err = client.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ = io.ReadAll(res.Body)
		if err = json.Unmarshal(body, &result); err != nil {
			panic(err)
		}
		scrollId = result.ScrollID

		if len(result.Classes) < 1 {
			// Giving up here, since it does not make sense to go any further
			break
		}

		ProcessClasses(result, config, client)
	}

	waitGroup.Wait()
}
