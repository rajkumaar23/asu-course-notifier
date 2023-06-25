package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func CheckAvailability(classNumber string) {
	defer WaitGroup.Done()

	courseCatalogURL := fmt.Sprintf(
		"%v?%v",
		CourseCatalogBaseURL,
		GetParamsForCourseCatalog(config, classNumber),
	)

	req, _ := http.NewRequest(http.MethodGet, courseCatalogURL, nil)
	req.Header.Set("Authorization", "Bearer null")

	res, err := HttpClient.Do(req)
	if err != nil {
		LogErrorAndPanic(errors.Join(errors.New("Error making request to get course catalog"), err))
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			LogErrorAndPanic(errors.Join(errors.New("Error closing request body of course catalog"), err))
		}
	}(res.Body)
	body, _ := io.ReadAll(res.Body)
	var result CourseCatalogResponse
	if err = json.Unmarshal(body, &result); err != nil {
		LogErrorAndPanic(errors.Join(errors.New("Error making request to get course catalog"), err))
	}

	if len(result.Classes) > 0 {
		ProcessClass(result.Classes[0])
	} else {
		fmt.Println("Length of result.Classes zero for", classNumber)
	}
}

func ProcessClass(class Class) {
	availableSlots := GetAvailableSlots(class)
	if availableSlots < 1 {
		fmt.Println("Slots unavailable for", class.Details.ClassNumber)
		return
	}

	fmt.Println("Slots available for", class.Details.ClassNumber)
	users := config.CoursesToWatch[class.Details.ClassNumber]
	for _, userName := range users {
		WaitGroup.Add(1)
		fmt.Println("Sending message to", userName, "for", class.Details.Title, "(", class.SubjectNumber, ")")
		go NotifyUser(userName, GetFormattedMessageForTelegram(userName, class, availableSlots), class)
	}
}

func NotifyUser(user string, message string, class Class) {
	defer WaitGroup.Done()

	telegramID, userExists := config.TelegramIDs[user]
	if !userExists {
		LogErrorAndPanic(errors.New("Telegram ID for " + user + " could not be found"))
	}

	params := url.Values{}
	params.Set("chat_id", strconv.Itoa(telegramID))
	params.Set("text", message)
	params.Set("parse_mode", "HTML")
	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage?%v", config.BotToken, params.Encode())

	req, _ := http.NewRequest(http.MethodGet, telegramURL, nil)
	res, err := HttpClient.Do(req)
	body, _ := io.ReadAll(res.Body)
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			LogErrorAndPanic(errors.Join(errors.New("Error closing request body when sending telegram message"), err))
		}
	}(res.Body)

	errorMessagePrefix := fmt.Sprintf("Sending message to %v for %v failed :", user, class.Details.Title)
	if res.StatusCode != 200 {
		LogErrorAndPanic(errors.New(errorMessagePrefix + string(body)))
	}
	if err != nil {
		LogErrorAndPanic(errors.Join(errors.New(errorMessagePrefix), err))
	}

	fmt.Println("Sent message to", user, "for", class.Details.Title, "(", class.SubjectNumber, ")")
}
