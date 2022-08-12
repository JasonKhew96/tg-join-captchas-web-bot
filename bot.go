package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func (cb *CaptchasBot) isValidChat(cjr *gotgbot.ChatJoinRequest) bool {
	chat := cb.getChatConfig(cjr.Chat.Id)
	return chat != nil
}

func (cb *CaptchasBot) timeoutKick(msgId int64, chat *gotgbot.Chat, user *gotgbot.User) func() {
	messages := cb.config.getMessages(user.LanguageCode)
	return func() {
		log.Println("timeout for user", chat.Id, user.Id, "message", msgId)
		if _, ok, err := cb.b.EditMessageText(messages.TimeoutError, &gotgbot.EditMessageTextOpts{
			ChatId:      user.Id,
			MessageId:   msgId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
		}); err != nil || !ok {
			log.Println("failed to edit message:", ok, err)
		}

		if _, err := cb.b.SendMessage(cb.config.LogChatId, buildLogString(&BuildLogStringParam{
			logType: LogTypeTimeout,
			chat:    chat,
			user:    user,
		}), &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		}); err != nil {
			log.Println("failed to send log message:", err)
		}
		cb.deleteStatusAndDecline(chat.Id, user.Id, false)
	}
}

func (cb *CaptchasBot) handleChatJoinRequest(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("ChatJoinRequest", ctx.EffectiveChat.Id, ctx.EffectiveSender.User.Id)

	if _, ok := cb.statusMap[ctx.EffectiveSender.User.Id]; ok {
		return nil
	}

	messages := cb.config.getMessages(ctx.EffectiveSender.User.LanguageCode)

	text := strings.Replace(messages.AskQuestion, `{chat_title}`, ctx.EffectiveChat.Title, -1)
	msg, err := b.SendMessage(ctx.EffectiveSender.User.Id, text, &gotgbot.SendMessageOpts{
		ProtectContent: true,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{gotgbot.InlineKeyboardButton{
					Text: messages.AskQuestionButton,
					WebApp: &gotgbot.WebAppInfo{
						Url: fmt.Sprintf("https://%s", cb.config.CustomDomain),
					},
				}},
			},
		},
	})
	if err != nil {
		return err
	}

	if _, err := b.SendMessage(cb.config.LogChatId, buildLogString(&BuildLogStringParam{
		logType: LogTypeRequested,
		chat:    ctx.EffectiveChat,
		user:    ctx.EffectiveSender.User,
		userBio: ctx.ChatJoinRequest.Bio,
	}), &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	}); err != nil {
		log.Println("failed to send log message:", err)
	}

	cb.statusMap[ctx.EffectiveSender.User.Id] = &Status{
		chat:      ctx.EffectiveChat,
		user:      ctx.EffectiveSender.User,
		msgId:     msg.MessageId,
		startTime: time.Now().Unix(),
		timer:     time.AfterFunc(time.Duration(cb.config.Timeout)*time.Second, cb.timeoutKick(msg.MessageId, ctx.EffectiveChat, ctx.EffectiveSender.User)),
	}
	return nil
}

func (cb *CaptchasBot) commandPing(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := ctx.EffectiveMessage.Reply(b, "pong", nil); err != nil {
		return err
	}
	return nil
}
