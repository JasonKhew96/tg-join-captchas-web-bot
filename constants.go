package main

const (
	LogFormatHeader          = "\\#%s\n"
	LogFormatChat            = "chat: `%s` \\[`%d`\\]\n"
	LogFormatUser            = "user: [%s](tg://user?id=%d) \\[`%d`\\]\n"
	LogFormatUsername        = "username: @%s\n"
	LogFormatBio             = "bio: `%s`\n"
	LogFormatIsGetChat       = "isGetChat: `%t`\n"
	LogFormatLanguage        = "language: `%s`\n"
	LogFormatPremium         = "premium: `%t`\n"
	LogFormatIsBlocked       = "blocked: `%t`\n"
	LogFormatData            = "data:\n`%s`\n"
	LogFormatAnswer          = "ID: %d\nA: %s\n"
	LogFormatVersion         = "version: `%s`\n"
	LogFormatPlatform        = "platform: `%s`\n"
	LogFormatStartTime       = "startTime: `%d`\n"
	LogFormatValidateElapsed = "validateElapsed: `%dms`\n"
	LogFormatSubmitElapsed   = "submitElapsed: `%dms`\n"
	LogFormatIp              = "ip: `%s`\n"
	LogFormatUserAgent       = "userAgent: `%s`\n"
)

type LogType int8

const (
	LogTypeRequested LogType = iota
	LogTypeApproved
	LogTypeTimeout
	LogTypeWrong
	LogTypeBanRegex
	LogTypeValidate
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
	case LogTypeBanRegex:
		return "BAN_REGEX"
	case LogTypeValidate:
		return "VALIDATE"
	default:
		return "UNKNOWN"
	}
}
