package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func parseUser(data string) (*User, error) {
	user := &User{}
	if err := json.Unmarshal([]byte(data), user); err != nil {
		return nil, err
	}
	return user, nil
}

func writeJson(w http.ResponseWriter, status bool, message string) {
	if respData, err := json.Marshal(CommonResponse{
		Status:  false,
		Message: message,
	}); err != nil {
		log.Println("failed to marshal response:", err.Error())
		w.Write([]byte(fmt.Sprintf(`{"status":false,"message":"%s"}`, message)))
	} else {
		w.Write(respData)
	}
}

func (cb *CaptchasBot) getChatConfig(chatId int64) *Chat {
	for _, chat := range cb.config.Chats {
		if chat.ChatId == chatId {
			return &chat
		}
	}
	return nil
}

func (cb *CaptchasBot) deleteStatusAndDecline(chatId, userId int64) {
	log.Println("Decline", chatId, userId)
	if userStatus, ok := cb.statusMap[userId]; ok {
		if _, err := cb.b.DeclineChatJoinRequest(chatId, userId, nil); err != nil {
			log.Println("failed to decline chat join request:", err)
		}
		if _, err := cb.b.BanChatMember(chatId, userId, &gotgbot.BanChatMemberOpts{
			UntilDate: time.Now().Unix() + cb.config.BanTime,
		}); err != nil {
			log.Println("failed to ban user:", err)
		}
		cb.stopStatusTimer(userStatus)
		delete(cb.statusMap, userId)
	}
}

func (cb *CaptchasBot) deleteStatusAndApprove(chatId, userId int64) {
	log.Println("Approve", chatId, userId)
	if userStatus, ok := cb.statusMap[userId]; ok {
		_, err := cb.b.ApproveChatJoinRequest(chatId, userId, nil)
		if err != nil {
			log.Println("failed to approve chat join request:", err)
		}
		cb.stopStatusTimer(userStatus)
		delete(cb.statusMap, userId)
	}
}

func (cb *CaptchasBot) stopStatusTimer(status *Status) {
	if !status.timer.Stop() {
		<-status.timer.C
	}
}
