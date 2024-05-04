package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/utils"
	"net/http"
)

// HandleResponse Handle response body
func HandleResponse(ctx *gin.Context, err error, data interface{}) {
	lang := utils.GetLang(ctx)
	trace, _ := ctx.Get(constant.TraceID)
	// no error
	if err == nil {
		ctx.JSON(http.StatusOK, NewRespBodyData(http.StatusOK, reason.Success, trace.(string), data).TrMsg(lang))
		return
	}

	ctx.JSON(http.StatusOK, NewRespBodyData(200, err.Error(), trace.(string), nil))
}

// BindAndCheck bind request and check
func BindAndCheck(ctx *gin.Context, data interface{}) bool {
	lang := utils.GetLang(ctx)
	ctx.Set(constant.AcceptLanguageFlag, lang)
	if err := ctx.ShouldBind(data); err != nil {
		glog.Slog.Errorf("http_handle BindAndCheck fail, %s", err.Error())
		HandleResponse(ctx, err, nil)
		return true
	}

	errField, err := validator.GetValidatorByLang(lang).Check(data)
	if err != nil {
		HandleResponse(ctx, err, errField)
		return true
	}
	return false
}

// BindAndCheckReturnErr bind request and check
func BindAndCheckReturnErr(ctx *gin.Context, data interface{}) (errFields []*validator.FormErrorField) {
	lang := utils.GetLang(ctx)
	if err := ctx.ShouldBind(data); err != nil {
		glog.Slog.Errorf("http_handle BindAndCheck fail, %s", err.Error())
		HandleResponse(ctx, err, nil)
		ctx.Abort()
		return nil
	}

	errFields, _ = validator.GetValidatorByLang(lang).Check(data)
	return errFields
}
