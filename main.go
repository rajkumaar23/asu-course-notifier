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
)

type CourseCatalogResponse struct {
	Classes []struct {
		Class struct {
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
	} `json:"classes"`
}

type Config struct {
	TelegramIDs    map[string]int      `json:"telegram_ids"`
	CoursesToWatch map[string][]string `json:"courses"`
	TermID         string              `json:"TERM_ID"`
	DepartmentCode string              `json:"DEPT_CODE"`
	BotToken       string              `json:"BOT_TOKEN"`
}

func GetParamsForCourseCatalog(config Config) string {
	params := url.Values{}
	params.Add("refine", "Y")
	params.Add("campusOrOnlineSelection", "A")
	params.Add("honors", "F")
	params.Add("level", "grad")
	params.Add("promod", "F")
	params.Add("searchType", "all")
	params.Add("subject", config.DepartmentCode)
	params.Add("term", config.TermID)

	return params.Encode()
}

func main() {
	configFileContents, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var config Config
	if err = json.Unmarshal(configFileContents, &config); err != nil {
		panic(err)
	}

	courseCatalogURL := fmt.Sprintf(
		"https://eadvs-cscc-catalog-api.apps.asu.edu/catalog-microservices/api/v1/search/classes?%v",
		GetParamsForCourseCatalog(config),
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

	for _, class := range result.Classes {
		reservedAvailability := 0
		var difference int
		for _, item := range class.ReservedSeatsInfo {
			difference = item.EnrollmentCap - item.EnrollmentTotal
			// To handle cases where ENRL_TOT > ENRL_CAP for some weird reason
			reservedAvailability += int(math.Max(float64(difference), 0))
		}

		availableSlots := class.SeatInfo.EnrollmentCap - class.SeatInfo.EnrollmentTotal - reservedAvailability
		if users, exists := config.CoursesToWatch[class.Class.ClassNumber]; exists && availableSlots > 0 {
			applyLink := fmt.Sprintf(
				"https://go.oasis.asu.edu/addclass/?STRM=%v&ASU_CLASS_NBR=%v",
				class.Class.Term,
				class.Class.ClassNumber,
			)
			swapLink := fmt.Sprintf("https://webapp4.asu.edu/myasu/?action=swapclass&strm=%v", config.TermID)
			for _, user := range users {
				telegramID := config.TelegramIDs[user]
				params := url.Values{}
				params.Add("chat_id", strconv.Itoa(telegramID))
				params.Add(
					"text",
					fmt.Sprintf(
						"Hey %v,\n\n%v (%v) by %v is available with %v slots.\n\nClick %v to add to cart.\n\nClick %v to swap with another course.",
						user,
						class.Class.Title,
						class.SubjectNumber,
						class.Class.Instructors[0],
						availableSlots,
						applyLink,
						swapLink,
					),
				)
				params.Add("parse_mode", "HTML")
				telegramURL := fmt.Sprintf("https://api.telegram.org/%v/sendMessage?%v", config.BotToken, params.Encode())
				req, _ = http.NewRequest(http.MethodGet, telegramURL, nil)
				_, err = client.Do(req)
				if err != nil {
					panic(err)
				}
				fmt.Println("Sent message to", user, "for", class.Class.Title, "(", class.SubjectNumber, ")")
			}
		}
	}
}
