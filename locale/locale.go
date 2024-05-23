package locale

import (
	"golang.org/x/text/language"
)

type I18nLanguage int

const (
	EnUS I18nLanguage = iota
	ZhCN
)

func ToLanguage(str string) I18nLanguage {
	switch str {
	case "zh-CN":
		return ZhCN
	case language.AmericanEnglish.String():
		return EnUS
	default:
		return ZhCN // default value
	}
}

type I18nMsgMap map[I18nLanguage]string

func (msgMap I18nMsgMap) SetMsg(lang I18nLanguage, msg string) I18nMsgMap {
	msgMap[lang] = msg
	return msgMap
}

// create new i18n message map, with default message
func NewI18nMsgMap() I18nMsgMap {
	return I18nMsgMap{}
}
