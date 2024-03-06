package locale

import "golang.org/x/text/language"

type I18nLanguage int

const (
	EnUS I18nLanguage = iota
	ZhCN
)

var SupportLanguages = []language.Tag{
	language.AmericanEnglish, // The first language is used as matcher fallback.
	language.Make("zh-CN"),
}

type LocaleInfo struct {
	Lang language.Tag
}

func DefaultLocaleInfo() LocaleInfo {
	return LocaleInfo{
		Lang: SupportLanguages[1],
	}
}

type I18nMsgMap map[string]string

func (msgMap I18nMsgMap) SetMsg(lang I18nLanguage, msg string) I18nMsgMap {
	msgMap[SupportLanguages[lang].String()] = msg
	return msgMap
}

// create new i18n message map, with default english message
func NewI18nMap(msg string) I18nMsgMap {
	return I18nMsgMap{
		SupportLanguages[0].String(): msg,
	}
}
