package main

const (
	LogFormatHeader   = "\\#%s\n"
	LogFormatChat     = "Chat: `%s` \\[`%d`\\]\n"
	LogFormatUser     = "User: [%s](tg://user?id=%d) \\[`%d`\\]\n"
	LogFormatUsername = "Username: @%s\n"
	LogFormatBio	  = "Bio: `%s`\n"
	LogFormatLanguage = "Language: `%s`\n"
	LogFormatPremium  = "Premium: `%t`\n"
	LogFormatData     = "Data:\n`%s`\n"
	LogFormatAnswer   = "ID: %d\nA: %s\n"
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
