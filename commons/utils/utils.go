package utils

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lawyer/commons/constant"
	"github.com/segmentfault/pacman/i18n"
)

// GetEnableShortID get language from header
func GetEnableShortID(ctx context.Context) bool {
	flag, ok := ctx.Value(constant.ShortIDFlag).(bool)
	if ok {
		return flag
	}
	return false
}

// GetLang get language from header
func GetLang(ctx *gin.Context) i18n.Language {
	acceptLanguage := ctx.GetHeader(constant.AcceptLanguageFlag)
	if len(acceptLanguage) == 0 {
		return i18n.DefaultLanguage
	}
	return i18n.Language(acceptLanguage)
}

func GetTraceId(ctx *gin.Context) string {
	trace := ctx.GetHeader(constant.TraceID)
	return trace
}

// GetLangByCtx get language from header
func GetLangByCtx(ctx context.Context) i18n.Language {
	acceptLanguage, ok := ctx.Value(constant.AcceptLanguageFlag).(i18n.Language)
	if ok {
		return acceptLanguage
	}
	return i18n.DefaultLanguage
}

func GenerateTraceId() string {
	newUUID, _ := uuid.NewUUID()
	return newUUID.String()
}
