package main

import (
	"log"
	"os"
	"regexp"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/liuzl/gocc"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("PORT is not set")
	}

	config, err := loadConfig()
	if err != nil {
		log.Fatal("failed to load config: ", err.Error())
		return
	}

	if config.BotToken == "" {
		log.Fatal("bot token is not set")
		return
	}

	banRegex := regexp.MustCompile(config.BanRegex)
	t2s, err := gocc.New("t2s")
	if err != nil {
		log.Fatal("failed to init t2s: ", err.Error())
		return
	}

	cb := &CaptchasBot{
		config:         config,
		statusMap:      make(map[int64]*Status),
		loggingChannel: make(chan MessageObject),
		banRegex:       banRegex,
		t2s:            t2s,
	}

	cb.runServer(port)

	b, err := gotgbot.NewBot(config.BotToken, &gotgbot.BotOpts{
		RequestOpts: &gotgbot.RequestOpts{
			Timeout: time.Minute,
			APIURL:  config.BotApiUrl,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	cb.b = b

	updater := ext.NewUpdater(nil)
	dispatcher := updater.Dispatcher

	go cb.telegramWorker(config.LogChatId, cb.loggingChannel)

	dispatcher.AddHandler(handlers.NewChatJoinRequest(cb.isValidChat, cb.handleChatJoinRequest))
	dispatcher.AddHandler(handlers.NewCommand("ping", cb.commandPing))

	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: false,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			AllowedUpdates: []string{"message", "callback_query", "chat_join_request"},
			Timeout:        60,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Minute,
				APIURL:  config.BotApiUrl,
			},
		},
	})
	if err != nil {
		log.Fatal("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", b.User.Username)

	updater.Idle()
}
