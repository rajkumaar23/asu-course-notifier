package main

import (
	"encoding/json"
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
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			PrintlnWithPrefixedTime("Error when closing body", err)
		}
	}(res.Body)
	body, _ := io.ReadAll(res.Body)
	var result CourseCatalogResponse
	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	if len(result.Classes) > 0 {
		ProcessClass(result.Classes[0])
	} else {
		PrintlnWithPrefixedTime("Length of result.Classes zero for", classNumber)
	}
}

func ProcessClass(class Class) {
	availableSlots := GetAvailableSlots(class)
	if availableSlots < 1 {
		PrintlnWithPrefixedTime("Slots unavailable for", class.Details.ClassNumber)
		return
	}

	PrintlnWithPrefixedTime("Slots available for", class.Details.ClassNumber)
	users := config.CoursesToWatch[class.Details.ClassNumber]
	for _, userName := range users {
		telegramID, userExists := config.TelegramIDs[userName]
		if !userExists {
			PrintlnWithPrefixedTime("Telegram ID for", userName, "could not be found")
			continue
		}

		WaitGroup.Add(1)
		PrintlnWithPrefixedTime("Sending message to", userName, "for", class.Details.Title, "(", class.SubjectNumber, ")")
		go NotifyUser(telegramID, userName, GetFormattedMessageForTelegram(userName, class, availableSlots), class)
	}
}

func NotifyUser(telegramID int, user string, message string, class Class) {
	defer WaitGroup.Done()

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
			PrintlnWithPrefixedTime("Error while closing the body", err)
		}
	}(res.Body)

	errorMessagePrefix := fmt.Sprintf("Sending message to %v for %v failed :", user, class.Details.Title)
	if res.StatusCode != 200 {
		PrintlnWithPrefixedTime(errorMessagePrefix, string(body))
		return
	}
	if err != nil {
		PrintlnWithPrefixedTime(errorMessagePrefix, err)
		return
	}

	PrintlnWithPrefixedTime("Sent message to", user, "for", class.Details.Title, "(", class.SubjectNumber, ")")
}
