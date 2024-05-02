package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/utils"
	"github.com/segmentfault/pacman/i18n"
)

var (
	langMapping = map[i18n.Language]bool{
		i18n.LanguageChinese:            true,
		i18n.LanguageChineseTraditional: true,
		i18n.LanguageEnglish:            true,
		i18n.LanguageGerman:             true,
		i18n.LanguageSpanish:            true,
		i18n.LanguageFrench:             true,
		i18n.LanguageItalian:            true,
		i18n.LanguageJapanese:           true,
		i18n.LanguageKorean:             true,
		i18n.LanguagePortuguese:         true,
		i18n.LanguageRussian:            true,
		i18n.LanguageVietnamese:         true,
	}
)

// ExtractAndSetAcceptLanguage extract accept language from header and set to context
func ExtractAndSetAcceptLanguage(ctx *gin.Context) {
	// The language of our front-end configuration, like en_US
	lang := utils.GetLang(ctx)
	if langMapping[lang] {
		ctx.Set(constant.AcceptLanguageFlag, lang)
	} else {
		// default language
		ctx.Set(constant.AcceptLanguageFlag, i18n.LanguageEnglish)
	}
	ctx.Next()
}
