package main

import (
	"os"

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

type Message struct {
	AskQuestionButton string `yaml:"ask_question_button"`
	AskQuestion       string `yaml:"ask_question"`
	TimeoutError      string `yaml:"timeout_error"`
	CorrectAnswer     string `yaml:"correct_answer"`
	WrongAnswer       string `yaml:"wrong_answer"`
}

type Config struct {
	BotToken     string `yaml:"bot_token"`
	BanTime      int64  `yaml:"ban_time"`
	Timeout      int64  `yaml:"timeout"`
	LogChatId    int64  `yaml:"log_chat_id"`
	CustomDomain string `yaml:"custom_domain"`
	DefaultLang  string `yaml:"default_lang"`
	Messages     map[string]Message
	Chats        []Chat `yaml:"chats"`
}

func loadConfig() (*Config, error) {
	c := Config{}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal([]byte(data), &c); err != nil {
		return nil, err
	}

	data, err = os.ReadFile("languages.yaml")
	if err != nil {
		return nil, err
	}

	var languages map[string]Message
	if err := yaml.Unmarshal([]byte(data), &languages); err != nil {
		return nil, err
	}

	c.Messages = languages

	return &c, nil
}

func (c *Config) getMessages(lang string) Message {
	messages, ok := c.Messages[lang]
	if !ok {
		return c.Messages[c.DefaultLang]
	}
	return messages
}
