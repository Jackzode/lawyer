package initServer

import (
	"github.com/apache/incubator-answer/internal/base/cron"
	"github.com/apache/incubator-answer/internal/base/server"
	"github.com/apache/incubator-answer/internal/base/translator"
	"github.com/apache/incubator-answer/internal/controller"
	templaterender "github.com/apache/incubator-answer/internal/controller/template_render"
	"github.com/apache/incubator-answer/internal/controller_admin"
	"github.com/apache/incubator-answer/internal/router"
	"github.com/apache/incubator-answer/internal/service/service_config"
	userexternallogin2 "github.com/apache/incubator-answer/internal/service/user_external_login"
	middleware2 "github.com/apache/incubator-answer/middleware"
	"github.com/segmentfault/pacman"
	"github.com/segmentfault/pacman/i18n"
)

var (
	I18nTranslator i18n.Translator
)

func initTranslator(i18nConf *translator.I18n) (err error) {
	I18nTranslator, err = translator.NewTranslator(i18nConf)
	return err
}

func initApplication(debug bool, serverConf *Server, serviceConf *service_config.ServiceConfig) (*pacman.Application, error) {

	langController := controller.NewLangController(I18nTranslator, siteInfoCommonService)
	staticRouter := router.NewStaticRouter(serviceConf)
	userController := controller.NewUserController(authService, userService, captchaService, emailService, siteInfoCommonService, userNotificationConfigService)
	rateLimitMiddleware := middleware2.NewRateLimitMiddleware(limitRepo)
	commentController := controller.NewCommentController(commentService, rankService, captchaService, rateLimitMiddleware)
	reportController := controller.NewReportController(reportService, rankService, captchaService)
	voteController := controller.NewVoteController(voteService, rankService, captchaService)
	tagController := controller.NewTagController(tagService, tagCommonService, rankService)
	followController := controller.NewFollowController(followService)
	collectionController := controller.NewCollectionController(collectionService)
	questionController := controller.NewQuestionController(questionService, answerService, rankService, siteInfoCommonService, captchaService, rateLimitMiddleware)
	answerController := controller.NewAnswerController(answerService, rankService, captchaService, siteInfoCommonService, rateLimitMiddleware)
	searchController := controller.NewSearchController(searchService, captchaService)
	revisionController := controller.NewRevisionController(serviceRevisionService, rankService)
	rankController := controller.NewRankController(rankService)
	controllerAdminReportController := controller_admin.NewReportController(reportAdminService)
	userAdminController := controller_admin.NewUserAdminController(userAdminService)
	reasonController := controller.NewReasonController(reasonService)
	themeController := controller_admin.NewThemeController()
	siteInfoController := controller_admin.NewSiteInfoController(siteInfoService)
	controllerSiteInfoController := controller.NewSiteInfoController(siteInfoCommonService)
	notificationController := controller.NewNotificationController(notificationService, rankService)
	dashboardController := controller.NewDashboardController(dashboardService)
	uploadController := controller.NewUploadController(uploaderService)
	activityController := controller.NewActivityController(activityService)
	roleController := controller_admin.NewRoleController(roleService)
	pluginController := controller_admin.NewPluginController(pluginCommonService)
	permissionController := controller.NewPermissionController(rankService)
	answerAPIRouter := router.NewAnswerAPIRouter(langController, userController, commentController, reportController, voteController, tagController, followController, collectionController, questionController, answerController, searchController, revisionController, rankController, controllerAdminReportController, userAdminController, reasonController, themeController, siteInfoController, controllerSiteInfoController, notificationController, dashboardController, uploadController, activityController, roleController, pluginController, permissionController)
	uiRouter := router.NewUIRouter(controllerSiteInfoController, siteInfoCommonService)
	authUserMiddleware := middleware2.NewAuthUserMiddleware(authService, siteInfoCommonService)
	avatarMiddleware := middleware2.NewAvatarMiddleware(serviceConf, uploaderService)
	shortIDMiddleware := middleware2.NewShortIDMiddleware(siteInfoCommonService)
	templateRenderController := templaterender.NewTemplateRenderController(questionService, userService, tagService, answerService, commentService, siteInfoCommonService, questionRepo)
	templateController := controller.NewTemplateController(templateRenderController, siteInfoCommonService)
	templateRouter := router.NewTemplateRouter(templateController, templateRenderController, siteInfoController, authUserMiddleware)
	connectorController := controller.NewConnectorController(siteInfoCommonService, emailService, userExternalLoginService)
	userCenterLoginService := userexternallogin2.NewUserCenterLoginService(userRepo, userCommon, userExternalLoginRepo, userActiveActivityRepo, siteInfoCommonService)
	userCenterController := controller.NewUserCenterController(userCenterLoginService, siteInfoCommonService)
	pluginAPIRouter := router.NewPluginAPIRouter(connectorController, userCenterController)
	ginEngine := server.NewHTTPServer(debug, staticRouter, answerAPIRouter, uiRouter, authUserMiddleware, avatarMiddleware, shortIDMiddleware, templateRouter, pluginAPIRouter)
	scheduledTaskManager := cron.NewScheduledTaskManager(siteInfoCommonService, questionService)
	application := newApplication(serverConf, ginEngine, scheduledTaskManager)
	return application, nil
}
