package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
	"github.com/lawyer/controller_admin"
	"github.com/lawyer/middleware"
)

// i18n
func RegisterLanguageApi(r *gin.RouterGroup) {
	c := &controller.LangController{}
	r.GET("/language/config", c.GetLangMapping)
	r.GET("/language/options", c.GetUserLangOptions)
	// language
	r.GET("/language/options", c.GetAdminLangOptions)

}

// siteinfo
func RegisterSiteInfoApi(r *gin.RouterGroup) {
	c := &controller.SiteInfoController{}
	r.GET("/siteinfo", c.GetSiteInfo)
	r.GET("/siteinfo/legal", c.GetSiteLegalInfo)

	ac := &controller_admin.SiteInfoController{}
	r.GET("/siteinfo/general", ac.GetGeneral)
	r.PUT("/siteinfo/general", ac.UpdateGeneral)
	r.GET("/siteinfo/interface", ac.GetInterface)
	r.PUT("/siteinfo/interface", ac.UpdateInterface)
	r.GET("/siteinfo/branding", ac.GetSiteBranding)
	r.PUT("/siteinfo/branding", ac.UpdateBranding)
	r.GET("/siteinfo/write", ac.GetSiteWrite)
	r.PUT("/siteinfo/write", ac.UpdateSiteWrite)
	r.GET("/siteinfo/legal", ac.GetSiteLegal)
	r.PUT("/siteinfo/legal", ac.UpdateSiteLegal)
	r.GET("/siteinfo/seo", ac.GetSeo)
	r.PUT("/siteinfo/seo", ac.UpdateSeo)
	r.GET("/siteinfo/login", ac.GetSiteLogin)
	r.PUT("/siteinfo/login", ac.UpdateSiteLogin)
	r.GET("/siteinfo/custom-css-html", ac.GetSiteCustomCssHTML)
	r.PUT("/siteinfo/custom-css-html", ac.UpdateSiteCustomCssHTML)
	r.GET("/siteinfo/theme", ac.GetSiteTheme)
	r.PUT("/siteinfo/theme", ac.SaveSiteTheme)
	r.GET("/siteinfo/users", ac.GetSiteUsers)
	r.PUT("/siteinfo/users", ac.UpdateSiteUsers)
	r.GET("/setting/smtp", ac.GetSMTPConfig)
	r.PUT("/setting/smtp", ac.UpdateSMTPConfig)
	r.GET("/setting/privileges", ac.GetPrivilegesConfig)
	r.PUT("/setting/privileges", ac.UpdatePrivilegesConfig)

}

func RegisterVoteApi(r *gin.RouterGroup) {
	// vote
	c := controller.NewVoteController()
	rg := r.Group("/vote", middleware.AccessToken())
	rg.POST("/up", c.VoteUp)
	rg.POST("/down", c.VoteDown)
	rg.GET("/personal/page", c.UserVotes)
}

func RegisterReportApi(r *gin.RouterGroup) {
	// report
	c := controller.NewReportController()
	r.POST("/report", c.AddReport)
	//管理员用接口
	ac := controller_admin.NewReportController()
	r.GET("/reports/page", ac.ListReportPage)
	r.PUT("/report", ac.Handle)
}

func RegisterOtherApi(r *gin.RouterGroup) {

	sc := controller.NewSearchController(nil, nil)
	r.GET("/search", sc.Search)
	r.GET("/search/desc", sc.SearchDesc)
	// rank
	rc := controller.NewRankController(nil)
	r.GET("/personal/rank/page", rc.GetRankPersonalWithPage)
	// follow
	fc := controller.NewFollowController(nil)
	r.POST("/follow", fc.Follow)
	r.PUT("/follow/tags", fc.UpdateFollowTags)
	// collection
	cc := controller.NewCollectionController(nil)
	r.POST("/collection/switch", cc.CollectionSwitch)
	// reason
	reasonC := controller.NewReasonController(nil)
	r.GET("/reasons", reasonC.Reasons)
	// activity
	acc := controller.NewActivityController(nil)
	r.GET("/activity/timeline", acc.GetObjectTimeline)
	r.GET("/activity/timeline/detail", acc.GetObjectTimelineDetail)
	// theme
	tc := controller_admin.NewThemeController()
	r.GET("/theme/options", tc.GetThemeOptions)
	// dashboard
	dc := controller.NewDashboardController(nil)
	r.GET("/dashboard", dc.DashboardInfo)
	// roles
	roleC := controller_admin.NewRoleController()
	r.GET("/roles", roleC.GetRoleList)
	// permission
	pc := controller.NewPermissionController(nil)
	r.GET("/permission", pc.GetPermission)
	// upload file
	uc := controller.NewUploadController()
	r.POST("/file", uc.UploadFile)
	r.POST("/post/render", uc.PostRender)

}
