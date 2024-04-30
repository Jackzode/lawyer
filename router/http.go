package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/middleware"
	"github.com/lawyer/router/routes"
)

// NewHTTPServer new http server.
func NewHTTPServer(debug bool) *gin.Engine {

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.GET("/heartBeat", heartBeats)
	InitRoutes(r)
	return r
}

func InitRoutes(r *gin.Engine) {

	//register middleware
	r.Use(middleware.ExtractAndSetAcceptLanguage)
	r.Use(middleware.TraceId)

	router := r.Group("/lawyer")
	// todo a lots of route....
	routes.RegisterUserApi(router)
	routes.RegisterOtherApi(router)
	routes.RegisterAnswerApi(router)
	routes.RegisterCommentApi(router)
	routes.RegisterNotificationApi(router)
	routes.RegisterReportApi(router)
	routes.RegisterTagApi(router)
	routes.RegisterRevisionApi(router)
	// plugin router
	//pluginAPIRouter.RegisterUnAuthConnectorRouter(adminauthV1)
	//pluginAPIRouter.RegisterAuthUserConnectorRouter(adminauthV1)
	//pluginAPIRouter.RegisterAuthAdminConnectorRouter(adminauthV1)
	//
	//_ = plugin.CallAgent(func(agent plugin.Agent) error {
	//	agent.RegisterUnAuthRouter(adminauthV1)
	//	agent.RegisterAuthUserRouter(adminauthV1)
	//	agent.RegisterAuthAdminRouter(adminauthV1)
	//	return nil
	//})

}

func heartBeats(ctx *gin.Context) { ctx.String(200, "OK") }
