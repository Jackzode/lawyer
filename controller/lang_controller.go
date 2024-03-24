package controller

import (
	"encoding/json"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/utils"
	services "github.com/lawyer/initServer/initServices"

	"github.com/gin-gonic/gin"
)

type LangController struct {
}

// NewLangController new language controller.
func NewLangController() *LangController {
	return &LangController{}
}

// GetLangMapping get language config mapping
// @Summary get language config mapping
// @Description get language config mapping
// @Tags Lang
// @Param Accept-Language header string true "Accept-Language"
// @Produce json
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/language/config [get]
func (u *LangController) GetLangMapping(ctx *gin.Context) {
	data, _ := services.I18nTranslator.Dump(utils.GetLang(ctx))
	var resp map[string]any
	_ = json.Unmarshal(data, &resp)
	handler.HandleResponse(ctx, nil, resp)
}

// GetAdminLangOptions Get language options
// @Summary Get language options
// @Description Get language options
// @Tags Lang
// @Produce json
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/language/options [get]
func (u *LangController) GetAdminLangOptions(ctx *gin.Context) {
	handler.HandleResponse(ctx, nil, translator.LanguageOptions)
}

// GetUserLangOptions Get language options
// @Summary Get language options
// @Description Get language options
// @Tags Lang
// @Produce json
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/language/options [get]
func (u *LangController) GetUserLangOptions(ctx *gin.Context) {
	siteInterfaceResp, err := services.SiteInfoCommonService.GetSiteInterface(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	options := translator.LanguageOptions
	if len(siteInterfaceResp.Language) > 0 {
		defaultOption := []*translator.LangOption{
			{Label: translator.DefaultLangOption, Value: translator.DefaultLangOption},
		}
		options = append(defaultOption, options...)
	}
	handler.HandleResponse(ctx, nil, options)
}
