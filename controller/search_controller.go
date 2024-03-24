package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	services "github.com/lawyer/initServer/initServices"
	"github.com/lawyer/middleware"
	"github.com/lawyer/plugin"
	"github.com/segmentfault/pacman/errors"
)

// SearchController tag controller
type SearchController struct {
}

// NewSearchController new controller
func NewSearchController() *SearchController {
	return &SearchController{}
}

// Search godoc
// @Summary search object
// @Description search object
// @Tags Search
// @Produce json
// @Security ApiKeyAuth
// @Param q query string true "query string"
// @Param order query string true "order" Enums(newest,active,score,relevance)
// @Success 200 {object} handler.RespBody{data=schema.SearchResp}
// @Router /answer/api/v1/search [get]
func (sc *SearchController) Search(ctx *gin.Context) {
	dto := schema.SearchDTO{}

	if handler.BindAndCheck(ctx, &dto) {
		return
	}
	dto.UserID = middleware.GetLoginUserIDFromContext(ctx)
	unit := ctx.ClientIP()
	if dto.UserID != "" {
		unit = dto.UserID
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := services.CaptchaService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionSearch, unit, dto.CaptchaID, dto.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	if !isAdmin {
		services.CaptchaService.ActionRecordAdd(ctx, entity.CaptchaActionSearch, unit)
	}
	resp, err := services.SearchService.Search(ctx, &dto)
	handler.HandleResponse(ctx, err, resp)
}

// SearchDesc get search description
// @Summary get search description
// @Description get search description
// @Tags Search
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SearchResp}
// @Router /answer/api/v1/search/desc [get]
func (sc *SearchController) SearchDesc(ctx *gin.Context) {
	var finder plugin.Search
	_ = plugin.CallSearch(func(search plugin.Search) error {
		finder = search
		return nil
	})
	resp := &schema.SearchDescResp{}
	if finder != nil {
		resp.Name = finder.Info().Name.Translate(ctx)
		resp.Icon = finder.Description().Icon
		resp.Link = finder.Description().Link
	}
	handler.HandleResponse(ctx, nil, resp)
}