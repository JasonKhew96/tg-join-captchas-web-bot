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
		if msgId != 0 {
			if _, ok, err := cb.b.EditMessageText(messages.TimeoutError, &gotgbot.EditMessageTextOpts{
				ChatId:      user.Id,
				MessageId:   msgId,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
			}); err != nil || !ok {
				log.Println("failed to edit message:", ok, err)
			}
		}

		cb.loggingChannel <- MessageObject{
			text: buildLogString(&BuildLogStringParam{
				logType: LogTypeTimeout,
				chat:    chat,
				user:    user,
			}),
			sendMessageOpts: &gotgbot.SendMessageOpts{
				ParseMode: "MarkdownV2",
			},
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
		log.Printf("failed to send request message: %s", err)
	}

	var isGetChat bool
	bio := ctx.ChatJoinRequest.Bio
	if bio == "" {
		chat, err := b.GetChat(ctx.EffectiveSender.User.Id, nil)
		if err != nil {
			log.Printf("failed to get chat: %s", err)
		} else if chat.Bio != "" {
			bio = chat.Bio
			isGetChat = true
		}
	}

	cb.loggingChannel <- MessageObject{
		text: buildLogString(&BuildLogStringParam{
			logType:     LogTypeRequested,
			chat:        ctx.EffectiveChat,
			user:        ctx.EffectiveSender.User,
			userBio:     bio,
			isGetChat:   isGetChat,
			isBlocked:   err != nil,
			startTimeMs: ctx.ChatJoinRequest.Date * 1e3,
		}),
		sendMessageOpts: &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		},
	}

	name := strings.Join([]string{ctx.EffectiveSender.User.FirstName, ctx.EffectiveSender.User.LastName}, " ")
	newName, err := cb.t2s.Convert(strings.Join([]string{ctx.EffectiveSender.User.FirstName, ctx.EffectiveSender.User.LastName}, " "))
	if err != nil {
		log.Println("failed to convert name:", err)
		newName = name
	}
	matchedName := cb.banRegex.MatchString(newName)
	newBio, err := cb.t2s.Convert(bio)
	if err != nil {
		log.Println("failed to convert bio:", err)
		newBio = bio
	}
	matchedBio := cb.banRegex.MatchString(newBio)
	if matchedName || matchedBio {
		log.Println("Regex ban", ctx.EffectiveChat.Id, ctx.EffectiveSender.User.Id)
		if _, err := b.DeclineChatJoinRequest(ctx.EffectiveChat.Id, ctx.EffectiveSender.User.Id, nil); err != nil {
			log.Println("failed to decline chat join request:", err)
		}
		if _, err := b.BanChatMember(ctx.EffectiveChat.Id, ctx.EffectiveSender.User.Id, nil); err != nil {
			log.Println("failed to ban user:", err)
		} else {
			cb.loggingChannel <- MessageObject{
				text: buildLogString(&BuildLogStringParam{
					logType:   LogTypeBanRegex,
					chat:      ctx.EffectiveChat,
					user:      ctx.EffectiveSender.User,
					userBio:   bio,
					isGetChat: isGetChat,
					isBlocked: err != nil,
				}),
				sendMessageOpts: &gotgbot.SendMessageOpts{
					ParseMode: "MarkdownV2",
				},
			}
		}
		return nil
	}

	var msgId int64
	if msg != nil {
		msgId = msg.MessageId
	}

	cb.statusMap[ctx.EffectiveSender.User.Id] = &Status{
		chat:        ctx.EffectiveChat,
		user:        ctx.EffectiveSender.User,
		msgId:       msgId,
		startTimeMs: time.Now().UnixMilli(),
		timer:       time.AfterFunc(time.Duration(cb.config.Timeout)*time.Second, cb.timeoutKick(msgId, ctx.EffectiveChat, ctx.EffectiveSender.User)),
	}
	return nil
}

func (cb *CaptchasBot) commandPing(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := ctx.EffectiveMessage.Reply(b, "pong", nil); err != nil {
		return err
	}
	return nil
}
