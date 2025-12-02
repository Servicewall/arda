package locale

type HttpResponseError interface {
	error
	MessageCode() *int
	MessageContent(I18nLanguage) string
	ResponseStatusCode() int
}

func NewAPIResponseError(statusCode int) *APIResponseError {
	return &APIResponseError{
		StatusCode: statusCode,
	}
}

type APIResponseError struct {
	StatusCode int        // HTTP 响应报文状态码
	BizCode    int        // 业务状态码(一般用不上)
	Message    string     // 代码调试信息
	MsgI18nMap I18nMsgMap // 接口返回的错误信息，支持国际化
}

func (err *APIResponseError) Error() string {
	if len(err.Message) > 0 {
		return err.Message
	}
	if msg, ok := err.MsgI18nMap[EnUS]; ok {
		return msg
	}
	return ""
}
func (err *APIResponseError) MessageCode() *int {
	if err.BizCode > 0 {
		return &err.BizCode
	}
	return nil
}
func (err *APIResponseError) MessageContent(lang I18nLanguage) string {
	if err.MsgI18nMap == nil {
		return err.Message
	}

	if i18nMsg, ok := err.MsgI18nMap[lang]; ok {
		return i18nMsg
	}

	return err.Message
}
func (err *APIResponseError) ResponseStatusCode() int {
	return err.StatusCode
}

func (err *APIResponseError) SetMessage(msg string) *APIResponseError {
	err.Message = msg
	return err
}

func (err *APIResponseError) SetMessageCode(code int) *APIResponseError {
	err.BizCode = code
	return err
}

func (err *APIResponseError) SetI18nMsg(lang I18nLanguage, msg string) *APIResponseError {
	if err.MsgI18nMap == nil {
		err.MsgI18nMap = I18nMsgMap{}
	}
	err.MsgI18nMap[lang] = msg
	return err
}
