package common

import "time"

const (
	ParamErrCode            int32 = 10000
	EmailRegisteredCode     int32 = 10001
	InternalErrorCode       int32 = 10002
	CaptchaErrCode          int32 = 10003
	UserAccountException    int32 = 10004
	UserAccountSuspended    int32 = 10005
	EmailRegisteredMsg            = "Email already registered"
	UserAccountExceptionMsg       = "user account exception"
	UserAccountSuspendedMsg       = "user account suspended"
	CaptchaErrMsg                 = "captcha is error"
	RequestParamErrMsg            = "request_param_error"
	ResponseErr                   = `{"code":10001, "errmsg":"internal error", "data":""}`
	OK                            = 0
	Success                       = "success"
	CaptchaPrefix                 = "user_"
	RegisterSuccessMsg            = "register success"

	TokenExpiration = 15 * 24 * time.Hour
	DefaultTraceId  = "NoTraceId"

	QuestionStatusAvailable int32 = 1
	QuestionUnPin           int32 = 1
	QuestionShow            int32 = 1
)
