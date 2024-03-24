package controller

import (
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/constant"
	services "github.com/lawyer/initServer/initServices"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/schema"
	"github.com/segmentfault/pacman/log"
)

type SiteInfoController struct {
	//siteInfoService siteinfo_common.SiteInfoCommonService
}

// NewSiteInfoController new site info controller.
func NewSiteInfoController() *SiteInfoController {
	return &SiteInfoController{}
}

// GetSiteInfo get site info
// @Summary get site info
// @Description get site info
// @Tags site
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteInfoResp}
// @Router /answer/api/v1/siteinfo [get]
func (sc *SiteInfoController) GetSiteInfo(ctx *gin.Context) {
	var err error
	resp := &schema.SiteInfoResp{Version: constant.Version, Revision: constant.Revision}
	resp.General, err = services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.Interface, err = services.SiteInfoCommonService.GetSiteInterface(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Branding, err = services.SiteInfoCommonService.GetSiteBranding(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Login, err = services.SiteInfoCommonService.GetSiteLogin(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Theme, err = services.SiteInfoCommonService.GetSiteTheme(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.CustomCssHtml, err = services.SiteInfoCommonService.GetSiteCustomCssHTML(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.SiteSeo, err = services.SiteInfoCommonService.GetSiteSeo(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.SiteUsers, err = services.SiteInfoCommonService.GetSiteUsers(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.Write, err = services.SiteInfoCommonService.GetSiteWrite(ctx)
	if err != nil {
		log.Error(err)
	}

	handler.HandleResponse(ctx, nil, resp)
}

// GetSiteLegalInfo get site legal info
// @Summary get site legal info
// @Description get site legal info
// @Tags site
// @Param info_type query string true "legal information type" Enums(tos, privacy)
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.GetSiteLegalInfoResp}
// @Router /answer/api/v1/siteinfo/legal [get]
func (sc *SiteInfoController) GetSiteLegalInfo(ctx *gin.Context) {
	req := &schema.GetSiteLegalInfoReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	siteLegal, err := services.SiteInfoCommonService.GetSiteLegal(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	resp := &schema.GetSiteLegalInfoResp{}
	if req.IsTOS() {
		resp.TermsOfServiceOriginalText = siteLegal.TermsOfServiceOriginalText
		resp.TermsOfServiceParsedText = siteLegal.TermsOfServiceParsedText
	} else if req.IsPrivacy() {
		resp.PrivacyPolicyOriginalText = siteLegal.PrivacyPolicyOriginalText
		resp.PrivacyPolicyParsedText = siteLegal.PrivacyPolicyParsedText
	}
	handler.HandleResponse(ctx, nil, resp)
}

// GetManifestJson get manifest.json
func (sc *SiteInfoController) GetManifestJson(ctx *gin.Context) {
	favicon := "favicon.ico"
	resp := &schema.GetManifestJsonResp{
		ManifestVersion: 3,
		Version:         constant.Version,
		Revision:        constant.Revision,
		ShortName:       "Answer",
		Name:            "answer.apache.org",
		Icons: map[string]string{
			"16":  favicon,
			"32":  favicon,
			"48":  favicon,
			"128": favicon,
		},
		StartUrl:        ".",
		Display:         "standalone",
		ThemeColor:      "#000000",
		BackgroundColor: "#ffffff",
	}
	branding, err := services.SiteInfoCommonService.GetSiteBranding(ctx)
	if err != nil {
		log.Error(err)
	} else if len(branding.Favicon) > 0 {
		resp.Icons["16"] = branding.Favicon
		resp.Icons["32"] = branding.Favicon
		resp.Icons["48"] = branding.Favicon
		resp.Icons["128"] = branding.Favicon
	}
	siteGeneral, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error(err)
	} else {
		resp.Name = siteGeneral.Name
		resp.ShortName = siteGeneral.Name
	}
	ctx.JSON(http.StatusOK, resp)
}