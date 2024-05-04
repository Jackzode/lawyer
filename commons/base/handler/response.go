package handler

import (
	"github.com/segmentfault/pacman/i18n"
)

// RespBody response body.
type RespBody struct {
	// http code
	Code int `json:"code"`
	//// reason key
	//Reason string `json:"reason"`
	//// response message
	Message string `json:"msg"`
	//
	TraceId string `json:"trace_id"`
	// response data
	Data interface{} `json:"data"`
}

// TrMsg translate the reason cause as a message
func (r *RespBody) TrMsg(lang i18n.Language) *RespBody {
	//if len(r.Message) == 0 {
	//	r.Message = translator.Tr(lang, r.Reason)
	//}
	return r
}

// NewRespBody new response body
//func NewRespBody(code int, reason string) *RespBody {
//	return &RespBody{
//		Code:   code,
//		Reason: reason,
//	}
//}

// NewRespBodyFromError new response body from error
//func NewRespBodyFromError(e error) *RespBody {
//	return &RespBody{
//		Code:    e.Code,
//		Reason:  e.Reason,
//		Message: e.Message,
//		Data: nil,
//	}
//}

// NewRespBodyData new response body with data
func NewRespBodyData(code int, msg, trace string, data interface{}) *RespBody {
	return &RespBody{
		Code:    code,
		Message: msg,
		TraceId: trace,
		Data:    data,
	}
}
