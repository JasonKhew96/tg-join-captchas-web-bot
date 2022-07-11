package main

import (
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Status struct {
	chatId    int64
	msgId int64
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

type SubmitData struct {
	CommonInitData
	Answers []struct {
		Id     int    `json:"id"`
		Answer string `json:"answer"`
	} `json:"answers"`
}

type CommonResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type ValidateResponse struct {
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
