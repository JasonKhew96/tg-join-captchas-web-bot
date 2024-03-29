package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/tomasen/realip"
)

func (cb *CaptchasBot) validate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
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

		if userStatus.validateTimeMs <= 0 {
			ip := realip.FromRequest(r)
			userStatus.ip = ip
			userStatus.userAgent = r.UserAgent()
			userStatus.validateTimeMs = time.Now().UnixMilli()
			cb.loggingChannel <- MessageObject{
				text: buildLogString(&BuildLogStringParam{
					logType:        LogTypeValidate,
					chat:           userStatus.chat,
					user:           userStatus.user,
					startTimeMs:    userStatus.startTimeMs,
					validateTimeMs: userStatus.validateTimeMs,
					ip:             ip,
					userAgent:      userStatus.userAgent,
				}),
				sendMessageOpts: &gotgbot.SendMessageOpts{
					ParseMode: "MarkdownV2",
				},
			}
		}

		chat := cb.getChatConfig(userStatus.chat.Id)

		if chat.ChatId == userStatus.chat.Id {
			var questions []QuestionResponse
			for _, question := range chat.Questions {
				questions = append(questions, QuestionResponse{
					Id:       question.Id,
					Type:     question.Type,
					Question: question.Question,
				})
			}
			if respData, err := json.Marshal(ValidateResponse{
				Title:          userStatus.chat.Title,
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
	body, err := io.ReadAll(r.Body)
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

		chat := cb.getChatConfig(userStatus.chat.Id)
		if chat == nil {
			writeJson(w, false, "validation failed")
			return
		}

		log.Println(chat.ChatId, user.Id, data.Version, data.Platform, data.Answers)

		if chat.ChatId == userStatus.chat.Id {
			correct := true
			for i := range data.Answers {
				for _, q := range chat.Questions {
					if data.Answers[i].Id == q.Id {
						if q.Type == "text" {
							re, err := regexp.Compile(q.Answer)
							if err != nil {
								log.Fatal(err)
							}
							if !re.MatchString(data.Answers[i].Answer) && correct {
								correct = false
								break
							}
						} else if q.Type == "hash" {
							if len(data.Answers[i].Answer) == 0 {
								correct = false
								break
							}
							decoded, err := base64.StdEncoding.DecodeString(data.Answers[i].Answer)
							if err != nil {
								correct = false
								break
							}
							newDecoded := string(decoded)
							if len(newDecoded) != 42 { // md5 32 + ts 10
								correct = false
								break
							}
							userHash := newDecoded[:32]
							userTimestampStr := newDecoded[32:]
							userTimestamp, err := strconv.ParseInt(userTimestampStr, 10, 64)
							if err != nil {
								correct = false
								break
							}
							if time.Now().UnixMilli()-(userTimestamp*1e3) > 300000 {
								correct = false
								break
							}
							expectedHashByte := md5.Sum([]byte(q.Answer + strconv.FormatInt(userTimestamp, 10)))
							expectedHash := hex.EncodeToString(expectedHashByte[:])
							if userHash != expectedHash {
								correct = false
								break
							}
						} else {
							correct = false
							break
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
			messages := cb.config.getMessages(userStatus.user.LanguageCode)
			if correct {
				if userStatus.msgId != 0 {
					if _, ok, err := cb.b.EditMessageText(messages.CorrectAnswer, &gotgbot.EditMessageTextOpts{
						ChatId:      user.Id,
						MessageId:   userStatus.msgId,
						ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
					}); err != nil || !ok {
						log.Println("failed to edit message:", ok, err)
					}
				}

				cb.loggingChannel <- MessageObject{
					text: buildLogString(&BuildLogStringParam{
						logType:        LogTypeApproved,
						chat:           userStatus.chat,
						user:           userStatus.user,
						answers:        data.Answers,
						version:        data.Version,
						platform:       data.Platform,
						startTimeMs:    userStatus.startTimeMs,
						validateTimeMs: userStatus.validateTimeMs,
						submitTimeMs:   time.Now().UnixMilli(),
						ip:             userStatus.ip,
						userAgent:      userStatus.userAgent,
					}),
					sendMessageOpts: &gotgbot.SendMessageOpts{
						ParseMode: "MarkdownV2",
					},
				}

				cb.deleteStatusAndApprove(userStatus.chat.Id, user.Id)
			} else {
				if userStatus.msgId != 0 {
					if _, ok, err := cb.b.EditMessageText(messages.WrongAnswer, &gotgbot.EditMessageTextOpts{
						ChatId:      user.Id,
						MessageId:   userStatus.msgId,
						ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
					}); err != nil || !ok {
						log.Println("failed to edit message:", ok, err)
					}
				}
				if _, err := cb.b.SendMessage(cb.config.LogChatId, buildLogString(&BuildLogStringParam{
					logType:        LogTypeWrong,
					chat:           userStatus.chat,
					user:           userStatus.user,
					answers:        data.Answers,
					version:        data.Version,
					platform:       data.Platform,
					startTimeMs:    userStatus.startTimeMs,
					validateTimeMs: userStatus.validateTimeMs,
					submitTimeMs:   time.Now().UnixMilli(),
					ip:             userStatus.ip,
					userAgent:      userStatus.userAgent,
				}), &gotgbot.SendMessageOpts{
					ParseMode: "MarkdownV2",
				}); err != nil {
					log.Println("failed to send log message:", err)
				}
				cb.deleteStatusAndDecline(userStatus.chat.Id, user.Id, true)
			}
			return
		}

		writeJson(w, false, "validation failed")
	} else {
		writeJson(w, false, "validation failed")
	}
}

func (cb *CaptchasBot) runServer(port string) {
	http.HandleFunc("/api/validate", cb.validate)
	http.HandleFunc("/api/submit", cb.submit)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Printf("Listening on 0.0.0.0:%s...", port)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
