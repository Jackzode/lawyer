package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/utils"
	myErrors "github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"net/http"
)

// HandleResponse Handle response body
func HandleResponse(ctx *gin.Context, err error, data interface{}) {
	lang := utils.GetLang(ctx)
	// no error
	if err == nil {
		ctx.JSON(http.StatusOK, NewRespBodyData(http.StatusOK, reason.Success, data).TrMsg(lang))
		return
	}

	var myErr *myErrors.Error
	// unknown error
	if !errors.As(err, &myErr) {
		log.Error(err, "\n", myErrors.LogStack(2, 5))
		ctx.JSON(http.StatusInternalServerError, NewRespBody(
			http.StatusInternalServerError, reason.UnknownError).TrMsg(lang))
		return
	}

	// log internal server error
	if myErrors.IsInternalServer(myErr) {
		log.Error(myErr)
	}

	respBody := NewRespBodyFromError(myErr).TrMsg(lang)
	if data != nil {
		respBody.Data = data
	}
	ctx.JSON(myErr.Code, respBody)
}

// BindAndCheck bind request and check
func BindAndCheck(ctx *gin.Context, data interface{}) bool {
	lang := utils.GetLang(ctx)
	ctx.Set(constant.AcceptLanguageFlag, lang)
	if err := ctx.ShouldBind(data); err != nil {
		log.Errorf("http_handle BindAndCheck fail, %s", err.Error())
		HandleResponse(ctx, myErrors.New(http.StatusBadRequest, reason.RequestFormatError), nil)
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
		log.Errorf("http_handle BindAndCheck fail, %s", err.Error())
		HandleResponse(ctx, myErrors.New(http.StatusBadRequest, reason.RequestFormatError), nil)
		ctx.Abort()
		return nil
	}

	errFields, _ = validator.GetValidatorByLang(lang).Check(data)
	return errFields
}