package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/service"
)

type LangController struct {
	//translator      i18n.Translator
	siteInfoService service.SiteInfoCommonService
}

// NewLangController new language controller.
func NewLangController(siteInfoService service.SiteInfoCommonService) *LangController {
	return &LangController{siteInfoService: siteInfoService}
}

func (u *LangController) GetLangMapping(ctx *gin.Context) {
	//data, _ := u.translator.Dump(utils.GetLang(ctx))
	//var resp map[string]any
	//_ = json.Unmarshal(data, &resp)
	//handler.HandleResponse(ctx, nil, resp)
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
	siteInterfaceResp, err := u.siteInfoService.GetSiteInterface(ctx)
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
