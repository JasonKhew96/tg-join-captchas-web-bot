package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func (cb *CaptchasBot) validate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Read body error: ", err)
		writeJson(w, false, "validation failed")
		return
	}
	data := &CommonInitData{}
	if err := json.Unmarshal(body, data); err != nil {
		log.Println("Unmarshal body error: ", err)
		writeJson(w, false, "validation failed")
		return
	}
	query, err := url.ParseQuery(data.InitData)
	if err != nil {
		log.Println("Parse body error: ", err)
		writeJson(w, false, "validation failed")
		return
	}
	ok, err := ext.ValidateWebAppQuery(query, cb.config.BotToken)
	if err != nil {
		log.Println("Validate body error: ", err)
		writeJson(w, false, "validation failed")
		return
	}
	if ok {
		user, err := parseUser(query.Get("user"))
		if err != nil {
			log.Println("Parse user error: ", err)
			writeJson(w, false, "validation failed")
			return
		}

		userStatus, isExists := cb.statusMap[user.Id]
		if !isExists {
			writeJson(w, false, "validation failed")
			return
		}

		chat := cb.getChatConfig(userStatus.chatId)

		if chat.ChatId == userStatus.chatId {
			var questions []QuestionResponse
			for _, question := range chat.Questions {
				questions = append(questions, QuestionResponse{
					Id:       question.Id,
					Type:     question.Type,
					Question: question.Question,
				})
			}
			if respData, err := json.Marshal(ValidateResponse{
				Title:          userStatus.title,
				Questions:      questions,
				CommonResponse: CommonResponse{Status: true, Message: "validation succeeded"},
			}); err != nil {
				log.Println("failed to marshal response:", err.Error())
				w.Write([]byte(`{"status":false,"message":"validation failed"}`))
			} else {
				w.Write(respData)
			}
			return
		}
		writeJson(w, false, "validation failed")
	} else {
		writeJson(w, false, "validation failed")
	}
}

func (cb *CaptchasBot) submit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Read body error: ", err)
		writeJson(w, false, "validation failed")
		return
	}
	data := &SubmitData{}
	if err := json.Unmarshal(body, data); err != nil {
		log.Println("Unmarshal body error: ", err)
		writeJson(w, false, "validation failed")
		return
	}
	query, err := url.ParseQuery(data.InitData)
	if err != nil {
		log.Println("Parse body error: ", err)
		writeJson(w, false, "validation failed")
		return
	}
	ok, err := ext.ValidateWebAppQuery(query, cb.config.BotToken)
	if err != nil {
		log.Println("Validate body error: ", err)
		writeJson(w, false, "validation failed")
		return
	}
	if ok {
		user, err := parseUser(query.Get("user"))
		if err != nil {
			log.Println("Parse user error: ", err)
			writeJson(w, false, "validation failed")
			return
		}

		userStatus, isExists := cb.statusMap[user.Id]
		if !isExists {
			writeJson(w, false, "validation failed")
			return
		}

		chat := cb.getChatConfig(userStatus.chatId)
		if chat == nil {
			writeJson(w, false, "validation failed")
			return
		}

		log.Println(chat.ChatId, user.Id, data.Answers)

		if chat.ChatId == userStatus.chatId {
			correct := true
			for _, userAnswer := range data.Answers {
				for _, q := range chat.Questions {
					if userAnswer.Id == q.Id {
						re, err := regexp.Compile(q.Answer)
						if err != nil {
							log.Fatal(err)
						}
						if !re.MatchString(userAnswer.Answer) && correct {
							correct = false
						}
					}
				}
			}
			if respData, err := json.Marshal(CommonResponse{
				Status: correct,
			}); err != nil {
				log.Println("failed to marshal response:", err.Error())
				w.Write([]byte(`{"status":false,"message":"validation failed"}`))
			} else {
				w.Write(respData)
			}
			if correct {
				if _, ok, err := cb.b.EditMessageText(cb.config.Messages.CorrectAnswer, &gotgbot.EditMessageTextOpts{
					ChatId:      user.Id,
					MessageId:   userStatus.msgId,
					ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
				}); err != nil || !ok {
					log.Println("failed to edit message:", ok, err)
				}
				cb.deleteStatusAndApprove(userStatus.chatId, user.Id)
			} else {
				if _, ok, err := cb.b.EditMessageText(cb.config.Messages.WrongAnswer, &gotgbot.EditMessageTextOpts{
					ChatId:      user.Id,
					MessageId:   userStatus.msgId,
					ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
				}); err != nil || !ok {
					log.Println("failed to edit message:", ok, err)
				}
				cb.deleteStatusAndDecline(userStatus.chatId, user.Id)
			}
			return
		}

		writeJson(w, false, "validation failed")
	} else {
		writeJson(w, false, "validation failed")
	}
}

func (cb *CaptchasBot) runServer(domain, port string) {
	http.HandleFunc("/api/validate", cb.validate)
	http.HandleFunc("/api/submit", cb.submit)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println(fmt.Sprintf("Listening on %s:%s...", domain, port))
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
