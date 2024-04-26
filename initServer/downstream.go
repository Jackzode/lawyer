package initServer

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/router"
)

// todo
func newApplication(server *gin.Engine) *gin.Engine {

	return server

}

func initApplication(debug bool) (*gin.Engine, error) {

	//langController := controller.NewLangController()
	//userController := controller.NewUserController()
	////todo
	////rateLimitMiddleware := middleware2.NewRateLimitMiddleware(repo.LimitRepo)
	//commentController := controller.NewCommentController()
	//reportController := controller.NewReportController()
	//voteController := controller.NewVoteController()
	//tagController := controller.NewTagController()
	//followController := controller.NewFollowController()
	//collectionController := controller.NewCollectionController()
	//questionController := controller.NewQuestionController()
	//answerController := controller.NewAnswerController()
	//searchController := controller.NewSearchController()
	//revisionController := controller.NewRevisionController()
	//rankController := controller.NewRankController()
	//controllerAdminReportController := controller_admin.NewReportController()
	//userAdminController := controller_admin.NewUserAdminController()
	//reasonController := controller.NewReasonController()
	//themeController := controller_admin.NewThemeController()
	//siteInfoController := controller_admin.NewSiteInfoController()
	//controllerSiteInfoController := controller.NewSiteInfoController()
	//notificationController := controller.NewNotificationController()
	//dashboardController := controller.NewDashboardController()
	//uploadController := controller.NewUploadController()
	//activityController := controller.NewActivityController()
	//roleController := controller_admin.NewRoleController()
	//pluginController := controller_admin.NewPluginController()
	//permissionController := controller.NewPermissionController()
	//answerAPIRouter := lawyerRoutes.NewAnswerAPIRouter(langController, userController, commentController, reportController, voteController, tagController, followController, collectionController, questionController, answerController, searchController, revisionController, rankController, controllerAdminReportController, userAdminController, reasonController, themeController, siteInfoController, controllerSiteInfoController, notificationController, dashboardController, uploadController, activityController, roleController, pluginController, permissionController)
	//uiRouter := router.NewUIRouter(controllerSiteInfoController, siteInfoCommonService)
	//authUserMiddleware := middleware2.NewAuthUserMiddleware(services.AuthService, services.SiteInfoCommonService)
	//avatarMiddleware := middleware2.NewAvatarMiddleware(services.UploaderService)
	//shortIDMiddleware := middleware2.NewShortIDMiddleware(services.SiteInfoCommonService)
	//todo
	//templateRenderController := templaterender.NewTemplateRenderController()
	//templateController := controller.NewTemplateController()
	//templateRouter := router.NewTemplateRouter(templateController, templateRenderController, siteInfoController, authUserMiddleware)
	//connectorController := controller.NewConnectorController()
	//userCenterLoginService := userexternallogin2.NewUserCenterLoginService(repo.UserRepo, services.UserCommon, repo.UserExternalLoginRepo, repo.UserActiveActivityRepo, services.SiteInfoCommonService)
	//userCenterController := controller.NewUserCenterController(userCenterLoginService)
	//pluginAPIRouter := plugin_api.NewPluginAPIRouter(connectorController, userCenterController)
	//
	ginEngine := router.NewHTTPServer(debug, nil)
	//scheduledTaskManager := cron.NewScheduledTaskManager(services.SiteInfoCommonService, services.QuestionService)
	//todo
	//application := newApplication(serverConf, ginEngine, scheduledTaskManager)
	//application := newApplication(ginEngine)

	return ginEngine, nil
}
