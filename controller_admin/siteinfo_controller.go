package controller_admin

import (
	"github.com/lawyer/commons/base/handler"
	services "github.com/lawyer/initServer/initServices"
	"github.com/lawyer/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/schema"
)

// SiteInfoController site info controller
type SiteInfoController struct {
	//siteInfoService *siteinfo.SiteInfoService
}

// NewSiteInfoController new site info controller
func NewSiteInfoController() *SiteInfoController {
	return &SiteInfoController{}
}

// GetGeneral get site general information
// @Summary get site general information
// @Description get site general information
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteGeneralResp}
// @Router /answer/admin/api/siteinfo/general [get]
func (sc *SiteInfoController) GetGeneral(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteGeneral(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetInterface get site interface
// @Summary get site interface
// @Description get site interface
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteInterfaceResp}
// @Router /answer/admin/api/siteinfo/interface [get]
func (sc *SiteInfoController) GetInterface(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteInterface(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteBranding get site interface
// @Summary get site interface
// @Description get site interface
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteBrandingResp}
// @Router /answer/admin/api/siteinfo/branding [get]
func (sc *SiteInfoController) GetSiteBranding(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteBranding(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteWrite get site interface
// @Summary get site interface
// @Description get site interface
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteWriteResp}
// @Router /answer/admin/api/siteinfo/write [get]
func (sc *SiteInfoController) GetSiteWrite(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteWrite(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteLegal Set the legal information for the site
// @Summary Set the legal information for the site
// @Description Set the legal information for the site
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteLegalResp}
// @Router /answer/admin/api/siteinfo/legal [get]
func (sc *SiteInfoController) GetSiteLegal(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteLegal(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSeo get site seo information
// @Summary get site seo information
// @Description get site seo information
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteSeoResp}
// @Router /answer/admin/api/siteinfo/seo [get]
func (sc *SiteInfoController) GetSeo(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSeo(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteLogin get site info login config
// @Summary get site info login config
// @Description get site info login config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteLoginResp}
// @Router /answer/admin/api/siteinfo/login [get]
func (sc *SiteInfoController) GetSiteLogin(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteLogin(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteCustomCssHTML get site info custom html css config
// @Summary get site info custom html css config
// @Description get site info custom html css config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteCustomCssHTMLResp}
// @Router /answer/admin/api/siteinfo/custom-css-html [get]
func (sc *SiteInfoController) GetSiteCustomCssHTML(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteCustomCssHTML(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteTheme get site info theme config
// @Summary get site info theme config
// @Description get site info theme config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteThemeResp}
// @Router /answer/admin/api/siteinfo/theme [get]
func (sc *SiteInfoController) GetSiteTheme(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteTheme(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteUsers get site user config
// @Summary get site user config
// @Description get site user config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteUsersResp}
// @Router /answer/admin/api/siteinfo/users [get]
func (sc *SiteInfoController) GetSiteUsers(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteUsers(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetRobots get site robots information
// @Summary get site robots information
// @Description get site robots information
// @Tags site
// @Produce json
// @Success 200 {string} txt ""
// @Router /robots.txt [get]
func (sc *SiteInfoController) GetRobots(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSeo(ctx)
	if err != nil {
		ctx.String(http.StatusOK, "")
		return
	}
	ctx.String(http.StatusOK, resp.Robots)
}

// GetRobots get site robots information
// @Summary get site robots information
// @Description get site robots information
// @Tags site
// @Produce json
// @Success 200 {string} txt ""
// @Router /custom.css [get]
func (sc *SiteInfoController) GetCss(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSiteCustomCssHTML(ctx)
	if err != nil {
		ctx.String(http.StatusOK, "")
		return
	}
	ctx.Header("content-type", "text/css;charset=utf-8")
	ctx.String(http.StatusOK, resp.CustomCss)
}

// UpdateSeo update site seo information
// @Summary update site seo information
// @Description update site seo information
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteSeoReq true "seo"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/seo [put]
func (sc *SiteInfoController) UpdateSeo(ctx *gin.Context) {
	req := schema.SiteSeoReq{}
	if handler.BindAndCheck(ctx, &req) {
		return
	}
	err := services.SiteInfoService.SaveSeo(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateGeneral update site general information
// @Summary update site general information
// @Description update site general information
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteGeneralReq true "general"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/general [put]
func (sc *SiteInfoController) UpdateGeneral(ctx *gin.Context) {
	req := schema.SiteGeneralReq{}
	if handler.BindAndCheck(ctx, &req) {
		return
	}
	err := services.SiteInfoService.SaveSiteGeneral(ctx, req)
	handler.HandleResponse(ctx, err, req)
}

// UpdateInterface update site interface
// @Summary update site info interface
// @Description update site info interface
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteInterfaceReq true "general"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/interface [put]
func (sc *SiteInfoController) UpdateInterface(ctx *gin.Context) {
	req := schema.SiteInterfaceReq{}
	if handler.BindAndCheck(ctx, &req) {
		return
	}
	err := services.SiteInfoService.SaveSiteInterface(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateBranding update site branding
// @Summary update site info branding
// @Description update site info branding
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteBrandingReq true "branding info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/branding [put]
func (sc *SiteInfoController) UpdateBranding(ctx *gin.Context) {
	req := &schema.SiteBrandingReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := services.SiteInfoService.SaveSiteBranding(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateSiteWrite update site write info
// @Summary update site write info
// @Description update site write info
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteWriteReq true "write info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/write [put]
func (sc *SiteInfoController) UpdateSiteWrite(ctx *gin.Context) {
	req := &schema.SiteWriteReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := services.SiteInfoService.SaveSiteWrite(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateSiteLegal update site legal info
// @Summary update site legal info
// @Description update site legal info
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteLegalReq true "write info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/legal [put]
func (sc *SiteInfoController) UpdateSiteLegal(ctx *gin.Context) {
	req := &schema.SiteLegalReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := services.SiteInfoService.SaveSiteLegal(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateSiteLogin update site login
// @Summary update site login
// @Description update site login
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteLoginReq true "login info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/login [put]
func (sc *SiteInfoController) UpdateSiteLogin(ctx *gin.Context) {
	req := &schema.SiteLoginReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := services.SiteInfoService.SaveSiteLogin(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateSiteCustomCssHTML update site custom css html config
// @Summary update site custom css html config
// @Description update site custom css html config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteCustomCssHTMLReq true "login info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/custom-css-html [put]
func (sc *SiteInfoController) UpdateSiteCustomCssHTML(ctx *gin.Context) {
	req := &schema.SiteCustomCssHTMLReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := services.SiteInfoService.SaveSiteCustomCssHTML(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// SaveSiteTheme update site custom css html config
// @Summary update site custom css html config
// @Description update site custom css html config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteThemeReq true "login info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/theme [put]
func (sc *SiteInfoController) SaveSiteTheme(ctx *gin.Context) {
	req := &schema.SiteThemeReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := services.SiteInfoService.SaveSiteTheme(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateSiteUsers update site config about users
// @Summary update site info config about users
// @Description update site info config about users
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteUsersReq true "users info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/users [put]
func (sc *SiteInfoController) UpdateSiteUsers(ctx *gin.Context) {
	req := &schema.SiteUsersReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := services.SiteInfoService.SaveSiteUsers(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetSMTPConfig get smtp config
// @Summary GetSMTPConfig get smtp config
// @Description GetSMTPConfig get smtp config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.GetSMTPConfigResp}
// @Router /answer/admin/api/setting/smtp [get]
func (sc *SiteInfoController) GetSMTPConfig(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetSMTPConfig(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateSMTPConfig update smtp config
// @Summary update smtp config
// @Description update smtp config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.UpdateSMTPConfigReq true "smtp config"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/setting/smtp [put]
func (sc *SiteInfoController) UpdateSMTPConfig(ctx *gin.Context) {
	req := &schema.UpdateSMTPConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := services.SiteInfoService.UpdateSMTPConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetPrivilegesConfig get privileges config
// @Summary GetPrivilegesConfig get privileges config
// @Description GetPrivilegesConfig get privileges config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.GetPrivilegesConfigResp}
// @Router /answer/admin/api/setting/privileges [get]
func (sc *SiteInfoController) GetPrivilegesConfig(ctx *gin.Context) {
	resp, err := services.SiteInfoService.GetPrivilegesConfig(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// UpdatePrivilegesConfig update privileges config
// @Summary update privileges config
// @Description update privileges config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.UpdatePrivilegesConfigReq true "config"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/setting/privileges [put]
func (sc *SiteInfoController) UpdatePrivilegesConfig(ctx *gin.Context) {
	req := &schema.UpdatePrivilegesConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := services.SiteInfoService.UpdatePrivilegesConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
