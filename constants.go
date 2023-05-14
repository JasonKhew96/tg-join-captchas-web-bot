package main

const (
	LogFormatHeader    = "\\#%s\n"
	LogFormatChat      = "chat: `%s` \\[`%d`\\]\n"
	LogFormatUser      = "user: [%s](tg://user?id=%d) \\[`%d`\\]\n"
	LogFormatUsername  = "username: @%s\n"
	LogFormatBio       = "bio: `%s`\n"
	LogFormatIsGetChat = "isGetChat: `%t`\n"
	LogFormatLanguage  = "language: `%s`\n"
	LogFormatPremium   = "premium: `%t`\n"
	LogFormatIsBlocked = "blocked: `%t`\n"
	LogFormatData      = "data:\n`%s`\n"
	LogFormatAnswer    = "ID: %d\nA: %s\n"
	LogFormatVersion   = "version: `%s`\n"
	LogFormatPlatform  = "platform: `%s`\n"
)

type LogType int8

const (
	LogTypeRequested LogType = iota
	LogTypeApproved
	LogTypeTimeout
	LogTypeWrong
)

func (t LogType) String() string {
	switch t {
	case LogTypeRequested:
		return "REQUESTED"
	case LogTypeApproved:
		return "APPROVED"
	case LogTypeTimeout:
		return "TIMEOUT"
	case LogTypeWrong:
		return "WRONG"
	default:
		return "UNKNOWN"
	}
}
