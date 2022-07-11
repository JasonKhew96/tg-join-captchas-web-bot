package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Question struct {
	Id       int    `yaml:"id"`
	Question string `yaml:"question"`
	Type     string `yaml:"type"`
	Answer   string `yaml:"answer"`
}

type Chat struct {
	ChatId    int64      `yaml:"chat_id"`
	Questions []Question `yaml:"questions"`
}

type Config struct {
	BotToken     string `yaml:"bot_token"`
	BanTime      int64  `yaml:"ban_time"`
	Timeout      int64  `yaml:"timeout"`
	CustomDomain string `yaml:"custom_domain"`
	Messages     struct {
		AskQuestionButton string `yaml:"ask_question_button"`
		AskQuestion       string `yaml:"ask_question"`
		TimeoutError      string `yaml:"timeout_error"`
		CorrectAnswer     string `yaml:"correct_answer"`
		WrongAnswer       string `yaml:"wrong_answer"`
	} `yaml:"messages"`
	Chats []Chat `yaml:"chats"`
}

func loadConfig() (*Config, error) {
	c := Config{}

	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
