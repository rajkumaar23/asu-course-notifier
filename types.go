package main

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
	BotToken       string              `json:"BOT_TOKEN"`
}
