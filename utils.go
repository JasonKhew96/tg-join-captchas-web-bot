package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

func (cb *CaptchasBot) deleteStatusAndDecline(chatId, userId int64, needBan bool) {
	log.Println("Decline", chatId, userId)
	if userStatus, ok := cb.statusMap[userId]; ok {
		if _, err := cb.b.DeclineChatJoinRequest(chatId, userId, nil); err != nil {
			log.Println("failed to decline chat join request:", err)
		}
		if needBan {
			if _, err := cb.b.BanChatMember(chatId, userId, &gotgbot.BanChatMemberOpts{
				UntilDate: time.Now().UnixMilli() + (cb.config.BanTime * 1e3),
			}); err != nil {
				log.Println("failed to ban user:", err)
			}
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
	status.timer.Stop()
	select {
	case <-status.timer.C:
	default:
	}
}

func buildLogString(param *BuildLogStringParam) string {
	var logString string
	logString += fmt.Sprintf(LogFormatHeader, EscapeMarkdownV2(param.logType.String()))
	logString += fmt.Sprintf(LogFormatChat, EscapeMarkdownV2(param.chat.Title), param.chat.Id)
	logString += fmt.Sprintf(LogFormatUser, EscapeMarkdownV2(strings.Join([]string{param.user.FirstName, param.user.LastName}, " ")), param.user.Id, param.user.Id)
	if param.user.Username != "" {
		logString += fmt.Sprintf(LogFormatUsername, EscapeMarkdownV2(param.user.Username))
	}
	if param.userBio != "" {
		logString += fmt.Sprintf(LogFormatBio, EscapeMarkdownV2(param.userBio))
	}
	if param.isGetChat {
		logString += fmt.Sprintf(LogFormatIsGetChat, param.isGetChat)
	}
	if param.user.LanguageCode != "" {
		logString += fmt.Sprintf(LogFormatLanguage, param.user.LanguageCode)
	}
	if param.user.IsPremium {
		logString += fmt.Sprintf(LogFormatPremium, param.user.IsPremium)
	}
	if param.isBlocked {
		logString += fmt.Sprintf(LogFormatIsBlocked, param.isBlocked)
	}
	if param.version != "" {
		logString += fmt.Sprintf(LogFormatVersion, param.version)
	}
	if param.platform != "" {
		logString += fmt.Sprintf(LogFormatPlatform, param.platform)
	}
	if param.submitTimeMs > 0 {
		logString += fmt.Sprintf(LogFormatValidateElapsed, param.validateTimeMs-param.startTimeMs)
		logString += fmt.Sprintf(LogFormatSubmitElapsed, param.submitTimeMs-param.validateTimeMs)
	} else if param.validateTimeMs > 0 {
		logString += fmt.Sprintf(LogFormatValidateElapsed, param.validateTimeMs-param.startTimeMs)
	} else if param.startTimeMs > 0 {
		logString += fmt.Sprintf(LogFormatStartTime, param.startTimeMs)
	}
	if param.ip != "" {
		logString += fmt.Sprintf(LogFormatIp, EscapeMarkdownV2(param.ip))
	}
	if param.userAgent != "" {
		logString += fmt.Sprintf(LogFormatUserAgent, EscapeMarkdownV2(param.userAgent))
	}
	if len(param.answers) > 0 {
		var data string
		for _, answer := range param.answers {
			data += fmt.Sprintf(LogFormatAnswer, answer.Id, EscapeMarkdownV2(answer.Answer))
		}
		logString += fmt.Sprintf(LogFormatData, data)
	}
	return logString
}

var allMdV2 = []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
var mdV2Repl = strings.NewReplacer(func() (out []string) {
	for _, x := range allMdV2 {
		out = append(out, x, "\\"+x)
	}
	return out
}()...)

func EscapeMarkdownV2(s string) string {
	return mdV2Repl.Replace(s)
}
