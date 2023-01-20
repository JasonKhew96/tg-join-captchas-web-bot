package main

import (
	"log"
	"os"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func main() {
	domain := os.Getenv("RAILWAY_STATIC_URL")
	port := os.Getenv("PORT")

	if domain == "" {
		log.Fatal("RAILWAY_STATIC_URL is not set")
	}
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

	cb := &CaptchasBot{
		domain:    domain,
		config:    config,
		statusMap: make(map[int64]*Status),
	}

	cb.runServer(domain, port)

	b, err := gotgbot.NewBot(config.BotToken, nil)
	if err != nil {
		log.Fatal(err)
	}

	cb.b = b

	updater := ext.NewUpdater(nil)
	dispatcher := updater.Dispatcher

	dispatcher.AddHandler(handlers.NewChatJoinRequest(cb.isValidChat, cb.handleChatJoinRequest))
	dispatcher.AddHandler(handlers.NewCommand("ping", cb.commandPing))

	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: false,
	})
	if err != nil {
		log.Fatal("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", b.User.Username)

	updater.Idle()
}
