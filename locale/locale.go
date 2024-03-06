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

func (lang I18nLanguage) Tag() string {
	return SupportLanguages[lang].String()
}

type I18nMsgMap map[string]string

func (msgMap I18nMsgMap) SetMsg(lang I18nLanguage, msg string) I18nMsgMap {
	msgMap[lang.Tag()] = msg
	return msgMap
}

// create new i18n message map, with default message
func NewI18nMsgMap() I18nMsgMap {
	return I18nMsgMap{}
}
