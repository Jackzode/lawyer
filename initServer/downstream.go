package initServer

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/cron"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/config"
	"github.com/lawyer/controller"
	templaterender "github.com/lawyer/controller/template_render"
	"github.com/lawyer/controller_admin"
	repo "github.com/lawyer/initServer/initRepo"
	services "github.com/lawyer/initServer/initServices"
	middleware2 "github.com/lawyer/middleware"
	"github.com/lawyer/router"
	userexternallogin2 "github.com/lawyer/service/user_external_login"
	"github.com/segmentfault/pacman/i18n"
)

var (
	I18nTranslator i18n.Translator
)

// todo
func initTranslator(i18nConf *config.I18n) (err error) {
	I18nTranslator, err = translator.NewTranslator(i18nConf)
	return err
}

// todo
func newApplication(serverConf *config.Server, server *gin.Engine, manager *cron.ScheduledTaskManager) *gin.Engine {

	return server

}

func initApplication(debug bool, serverConf *config.Server) (*gin.Engine, error) {

	langController := controller.NewLangController(I18nTranslator, services.SiteInfoCommonService)
	userController := controller.NewUserController(services.AuthService, services.UserService, services.CaptchaService, services.EmailService, services.SiteInfoCommonService, services.UserNotificationConfigService)
	rateLimitMiddleware := middleware2.NewRateLimitMiddleware(repo.LimitRepo)
	commentController := controller.NewCommentController(services.CommentService, services.RankService, services.CaptchaService, rateLimitMiddleware)
	reportController := controller.NewReportController(services.ReportService, services.RankService, services.CaptchaService)
	voteController := controller.NewVoteController(services.VoteService, services.RankService, services.CaptchaService)
	tagController := controller.NewTagController(services.TagService, services.TagCommonService, services.RankService)
	followController := controller.NewFollowController(services.FollowService)
	collectionController := controller.NewCollectionController(services.CollectionService)
	questionController := controller.NewQuestionController(services.QuestionService, services.AnswerService, services.RankService, services.SiteInfoCommonService, services.CaptchaService, rateLimitMiddleware)
	answerController := controller.NewAnswerController(services.AnswerService, services.RankService, services.CaptchaService, services.SiteInfoCommonService, rateLimitMiddleware)
	searchController := controller.NewSearchController(services.SearchService, services.CaptchaService)
	revisionController := controller.NewRevisionController(services.ServiceRevisionService, services.RankService)
	rankController := controller.NewRankController(services.RankService)
	controllerAdminReportController := controller_admin.NewReportController(services.ReportAdminService)
	userAdminController := controller_admin.NewUserAdminController(services.UserAdminService)
	reasonController := controller.NewReasonController(services.ReasonService)
	themeController := controller_admin.NewThemeController()
	siteInfoController := controller_admin.NewSiteInfoController(services.SiteInfoService)
	controllerSiteInfoController := controller.NewSiteInfoController(services.SiteInfoCommonService)
	notificationController := controller.NewNotificationController(services.NotificationService, services.RankService)
	dashboardController := controller.NewDashboardController(services.DashboardService)
	uploadController := controller.NewUploadController(services.UploaderService)
	activityController := controller.NewActivityController(services.ActivityService)
	roleController := controller_admin.NewRoleController(services.RoleService)
	pluginController := controller_admin.NewPluginController(services.PluginCommonService)
	permissionController := controller.NewPermissionController(services.RankService)
	answerAPIRouter := router.NewAnswerAPIRouter(langController, userController, commentController, reportController, voteController, tagController, followController, collectionController, questionController, answerController, searchController, revisionController, rankController, controllerAdminReportController, userAdminController, reasonController, themeController, siteInfoController, controllerSiteInfoController, notificationController, dashboardController, uploadController, activityController, roleController, pluginController, permissionController)
	//uiRouter := router.NewUIRouter(controllerSiteInfoController, siteInfoCommonService)
	authUserMiddleware := middleware2.NewAuthUserMiddleware(services.AuthService, services.SiteInfoCommonService)
	avatarMiddleware := middleware2.NewAvatarMiddleware(services.UploaderService)
	shortIDMiddleware := middleware2.NewShortIDMiddleware(services.SiteInfoCommonService)
	templateRenderController := templaterender.NewTemplateRenderController(services.QuestionService, services.UserService, services.TagService, services.AnswerService, services.CommentService, services.SiteInfoCommonService, repo.QuestionRepo)
	templateController := controller.NewTemplateController(templateRenderController, services.SiteInfoCommonService)
	templateRouter := router.NewTemplateRouter(templateController, templateRenderController, siteInfoController, authUserMiddleware)
	connectorController := controller.NewConnectorController(services.SiteInfoCommonService, services.EmailService, services.UserExternalLoginService)
	userCenterLoginService := userexternallogin2.NewUserCenterLoginService(repo.UserRepo, services.UserCommon, repo.UserExternalLoginRepo, repo.UserActiveActivityRepo, services.SiteInfoCommonService)
	userCenterController := controller.NewUserCenterController(userCenterLoginService, services.SiteInfoCommonService)
	pluginAPIRouter := router.NewPluginAPIRouter(connectorController, userCenterController)
	//
	ginEngine := router.NewHTTPServer(debug, answerAPIRouter, authUserMiddleware, avatarMiddleware, shortIDMiddleware, templateRouter, pluginAPIRouter)
	scheduledTaskManager := cron.NewScheduledTaskManager(services.SiteInfoCommonService, services.QuestionService)
	//todo
	application := newApplication(serverConf, ginEngine, scheduledTaskManager)
	return application, nil
}
