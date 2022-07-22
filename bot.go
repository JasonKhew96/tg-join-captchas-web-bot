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

func (cb *CaptchasBot) timeoutBan(chatId, userId, msgId int64, lang string) func() {
	messages := cb.config.getMessages(lang)
	return func() {
		log.Println("timeout for user", chatId, userId, "message", msgId)
		if _, ok, err := cb.b.EditMessageText(messages.TimeoutError, &gotgbot.EditMessageTextOpts{
			ChatId:      userId,
			MessageId:   msgId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
		}); err != nil || !ok {
			log.Println("failed to edit message:", ok, err)
		}
		cb.deleteStatusAndDecline(chatId, userId)
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

	cb.statusMap[ctx.EffectiveSender.User.Id] = &Status{
		title:     ctx.EffectiveChat.Title,
		lang:      ctx.EffectiveSender.User.LanguageCode,
		chatId:    ctx.EffectiveChat.Id,
		msgId:     msg.MessageId,
		startTime: time.Now().Unix(),
		timer:     time.AfterFunc(time.Duration(cb.config.Timeout)*time.Second, cb.timeoutBan(ctx.EffectiveChat.Id, ctx.EffectiveSender.User.Id, msg.MessageId, ctx.EffectiveSender.User.LanguageCode)),
	}
	return nil
}

func (cb *CaptchasBot) commandPing(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := ctx.EffectiveMessage.Reply(b, "pong", nil); err != nil {
		return err
	}
	return nil
}
