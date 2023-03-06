package main

import (
	"fmt"
	"math"
	"net/url"
	"time"
)

func GetParamsForCourseCatalog(config Config, classNumber string) string {
	params := url.Values{}
	params.Add("refine", "Y")
	params.Add("term", config.TermID)
	params.Add("keywords", classNumber)
	return params.Encode()
}

func GetFormattedMessageForTelegram(user string, class Class, availableSlots int) string {
	applyLink := fmt.Sprintf(
		"https://go.oasis.asu.edu/addclass/?STRM=%v&ASU_CLASS_NBR=%v",
		class.Details.Term,
		class.Details.ClassNumber,
	)
	swapLink := fmt.Sprintf("https://webapp4.asu.edu/myasu/?action=swapclass&strm=%v", config.TermID)

	var instructor string
	if len(class.Details.Instructors) > 0 {
		instructor = class.Details.Instructors[0]
	} else {
		instructor = "Staff"
	}

	return fmt.Sprintf(
		"Hey %v,\n\n<b>%v (%v) by %v is available with %v slots</b>.\n\nClick %v to add to cart.\n\nClick %v to swap with another course.",
		user,
		class.Details.Title,
		class.Details.ClassNumber,
		instructor,
		availableSlots,
		applyLink,
		swapLink,
	)
}

func GetAvailableSlots(class Class) int {
	reservedAvailability := 0
	var difference int
	for _, item := range class.ReservedSeatsInfo {
		difference = item.EnrollmentCap - item.EnrollmentTotal
		// To handle cases where ENRL_TOT > ENRL_CAP for some weird reason
		reservedAvailability += int(math.Max(float64(difference), 0))
	}

	return class.SeatInfo.EnrollmentCap - class.SeatInfo.EnrollmentTotal - reservedAvailability
}

func GetFormattedCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func PrintlnWithPrefixedTime(message ...any) {
	fmt.Println(GetFormattedCurrentTime(), ":", message)
}
