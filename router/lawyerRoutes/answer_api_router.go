package lawyerRoutes

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

func RegisterQuestionApi(r *gin.RouterGroup) {

	c := controller.QuestionController{}
	// question
	r.GET("/question/info", c.GetQuestion)
	r.GET("/question/invite", c.GetQuestionInviteUserInfo)
	r.GET("/question/page", c.QuestionPage)
	r.GET("/question/similar/tag", c.SimilarQuestion)
	r.GET("/personal/qa/top", c.UserTop)
	r.GET("/personal/question/page", c.PersonalQuestionPage)
	r.GET("/answer/page", c.AdminAnswerPage)
	// question
	r.GET("/personal/collection/page", c.PersonalCollectionPage)
	r.POST("/question", c.AddQuestion)
	r.POST("/question/answer", c.AddQuestionByAnswer)
	r.PUT("/question", c.UpdateQuestion)
	r.PUT("/question/invite", c.UpdateQuestionInviteUser)
	r.DELETE("/question", c.RemoveQuestion)
	r.PUT("/question/status", c.CloseQuestion)
	r.PUT("/question/operation", c.OperationQuestion)
	r.PUT("/question/reopen", c.ReopenQuestion)
	r.GET("/question/similar", c.GetSimilarQuestions)
	r.POST("/question/recover", c.QuestionRecover)
	r.GET("/question/page", c.AdminQuestionPage)
	r.PUT("/question/status", c.AdminUpdateQuestionStatus)
	r.GET("/personal/answer/page", c.PersonalAnswerPage)

}

func RegisterAnswerApi(r *gin.RouterGroup) {

	c := &controller.AnswerController{}
	// answer
	r.GET("/answer/info", c.Get)
	r.GET("/answer/page", c.AnswerList)
	r.POST("/answer", c.Add)
	r.PUT("/answer", c.Update)
	r.POST("/answer/acceptance", c.Accepted)
	r.DELETE("/answer", c.RemoveAnswer)
	r.POST("/answer/recover", c.RecoverAnswer)
	r.PUT("/answer/status", c.AdminUpdateAnswerStatus)

}

func RegisterUserApi(r *gin.RouterGroup) {

	c := controller.NewUserController()
	ac := controller_admin.NewUserAdminController()
	r.GET("/user/info", c.GetUserInfoByUserID) //need login
	r.GET("/user/action/record", c.ActionRecord)
	//skip plugin
	routerGroup := r.Group("", middleware.BanAPIForUserCenter)

	routerGroup.GET("/user/register/captcha", c.UserRegisterCaptcha)
	routerGroup.POST("/user/register/email", c.UserRegisterByEmail)
	routerGroup.POST("/user/email/verification", c.UserVerifyEmail)
	routerGroup.POST("/user/login/email", c.UserEmailLogin)
	routerGroup.PUT("/user/email", c.UserChangeEmailVerify)
	routerGroup.POST("/user/password/reset", c.RetrievePassWord)
	routerGroup.POST("/user/password/replacement", c.UseRePassWord)
	routerGroup.PUT("/user/notification/unsubscribe", c.UserUnsubscribeNotification)
	// user
	r.GET("/user/logout", c.UserLogout)
	r.POST("/user/email/change/code", middleware.BanAPIForUserCenter, c.UserChangeEmailSendCode)
	r.POST("/user/email/verification/send", middleware.BanAPIForUserCenter, c.UserVerifyEmailSend)
	r.GET("/personal/user/info", c.GetOtherUserInfoByUsername)
	r.GET("/user/ranking", c.UserRanking)

	// user
	r.GET("/users/page", ac.GetUserPage)
	r.PUT("/user/status", ac.UpdateUserStatus)
	r.PUT("/user/role", ac.UpdateUserRole)
	r.GET("/user/activation", ac.GetUserActivation)
	r.POST("/user/activation", ac.SendUserActivation)
	r.POST("/user", ac.AddUser)
	r.POST("/users", ac.AddUsers)
	r.PUT("/user/password", ac.UpdateUserPassword)
	r.PUT("/user/password", middleware.BanAPIForUserCenter, c.UserModifyPassWord)
	r.PUT("/user/info", c.UserUpdateInfo)
	r.PUT("/user/interface", c.UserUpdateInterface)
	r.GET("/user/notification/config", c.GetUserNotificationConfig)
	r.PUT("/user/notification/config", c.UpdateUserNotificationConfig)
	r.GET("/user/info/search", c.SearchUserListByName)

	//r.GET("/user/action/record", authUserMiddleware.Auth(), c.ActionRecord)
}

// revision
func RegisterRevisionApi(r *gin.RouterGroup) {
	c := controller.NewRevisionController(nil, nil)
	r.GET("/revisions", c.GetRevisionList)
	r.GET("/revisions/unreviewed", c.GetUnreviewedRevisionList)
	r.PUT("/revisions/audit", c.RevisionAudit)
	r.GET("/revisions/edit/check", c.CheckCanUpdateRevision)

}

func RegisterTagApi(r *gin.RouterGroup) {
	c := controller.NewTagController(nil, nil, nil)
	// tag
	r.GET("/tags/page", c.GetTagWithPage)
	r.GET("/tags/following", c.GetFollowingTags)
	r.GET("/tag", c.GetTagInfo)
	r.GET("/tags", c.GetTagsBySlugName)
	r.GET("/tag/synonyms", c.GetTagSynonyms)
	// tag
	r.GET("/question/tags", c.SearchTagLike)
	r.POST("/tag", c.AddTag)
	r.PUT("/tag", c.UpdateTag)
	r.POST("/tag/recover", c.RecoverTag)
	r.DELETE("/tag", c.RemoveTag)
	r.PUT("/tag/synonym", c.UpdateTagSynonym)
}

func RegisterCommentApi(r *gin.RouterGroup) {
	c := &controller.CommentController{}
	// comment
	r.POST("/comment", c.AddComment)
	r.DELETE("/comment", c.RemoveComment)
	r.PUT("/comment", c.UpdateComment)
	r.GET("/comment/page", c.GetCommentWithPage)
	r.GET("/personal/comment/page", c.GetCommentPersonalWithPage)
	r.GET("/comment", c.GetComment)
}

func RegisterNotificationApi(r *gin.RouterGroup) {
	c := controller.NewNotificationController(nil, nil)
	// notification
	r.GET("/notification/status", c.GetRedDot)
	r.PUT("/notification/status", c.ClearRedDot)
	r.GET("/notification/page", c.GetList)
	r.PUT("/notification/read/state/all", c.ClearUnRead)
	r.PUT("/notification/read/state", c.ClearIDUnRead)
}

func RegisterVoteApi(r *gin.RouterGroup) {
	// vote
	c := controller.NewVoteController()
	r.GET("/personal/vote/page", c.UserVotes)
	r.POST("/vote/up", c.VoteUp)
	r.POST("/vote/down", c.VoteDown)
}

func RegisterReportApi(r *gin.RouterGroup) {
	// report
	c := controller.NewReportController(nil, nil, nil)
	r.POST("/report", c.AddReport)
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

//func (a *AnswerAPIRouter) RegisterAnswerAPIRouter(r *gin.RouterGroup) {
//
//
//}
