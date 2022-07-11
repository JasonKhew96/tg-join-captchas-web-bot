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

func (cb *CaptchasBot) timeoutBan(chatId, userId, msgId int64) func() {
	return func() {
		log.Println("timeout for user", chatId, userId, "message", msgId)
		if _, ok, err := cb.b.EditMessageText(cb.config.Messages.TimeoutError, &gotgbot.EditMessageTextOpts{
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
	log.Println("ChatJoinRequest", ctx.EffectiveChat.Id, ctx.EffectiveUser.Id)
	text := strings.Replace(cb.config.Messages.AskQuestion, `{chat_title}`, ctx.EffectiveChat.Title, -1)
	msg, err := b.SendMessage(ctx.EffectiveUser.Id, text, &gotgbot.SendMessageOpts{
		ProtectContent: true,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{gotgbot.InlineKeyboardButton{
					Text: cb.config.Messages.AskQuestionButton,
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
	cb.statusMap[ctx.EffectiveUser.Id] = &Status{
		chatId:    ctx.EffectiveChat.Id,
		msgId:     msg.MessageId,
		startTime: time.Now().Unix(),
		timer:     time.AfterFunc(time.Duration(cb.config.Timeout)*time.Second, cb.timeoutBan(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id, msg.MessageId)),
	}
	return nil
}

func (cb *CaptchasBot) commandPing(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := ctx.EffectiveMessage.Reply(b, "pong", nil); err != nil {
		return err
	}
	return nil
}
