package main

import (
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type MessageObject struct {
	text            string
	sendMessageOpts *gotgbot.SendMessageOpts
}

func (cb *CaptchasBot) work(chatId int64, message MessageObject) {
	_, err := cb.b.SendMessage(chatId, message.text, message.sendMessageOpts)
	if err != nil {
		log.Println("failed to send message:", message, err)
	}
}

func (cb *CaptchasBot) telegramWorker(chatId int64, messages <-chan MessageObject) {
	for message := range messages {
		cb.work(chatId, message)
		time.Sleep(3 * time.Second)
	}
}
