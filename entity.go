package main

import (
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Status struct {
	chat      *gotgbot.Chat
	user      *gotgbot.User
	msgId     int64
	startTime int64
	timer     *time.Timer
}

type CaptchasBot struct {
	domain    string
	config    *Config
	b         *gotgbot.Bot
	statusMap map[int64]*Status
}

type CommonInitData struct {
	InitData string `json:"init_data"`
}

type Answer struct {
	Id     int    `json:"id"`
	Answer string `json:"answer"`
}

type SubmitData struct {
	CommonInitData
	Answers []Answer `json:"answers"`
}

type CommonResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type ValidateResponse struct {
	Title     string             `json:"title"`
	Questions []QuestionResponse `json:"questions"`
	CommonResponse
}

type QuestionResponse struct {
	Id       int    `json:"id"`
	Question string `json:"question"`
	Type     string `json:"type"`
}

type User struct {
	Id           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type BuildLogStringParam struct {
	logType LogType
	chat    *gotgbot.Chat
	user    *gotgbot.User
	userBio string
	answers []Answer
}
