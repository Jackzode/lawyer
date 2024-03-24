package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/middleware"
	"github.com/lawyer/router/lawyerRoutes"
	p "github.com/lawyer/router/plugin_api"
)

// NewHTTPServer new http server.
func NewHTTPServer(debug bool, pluginAPIRouter *p.PluginAPIRouter) *gin.Engine {

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.GET("/heartBeat", heartBeat)
	InitRoutes(r, pluginAPIRouter)
	return r
}

func InitRoutes(r *gin.Engine, pluginAPIRouter *p.PluginAPIRouter) {

	//register middleware
	r.Use(middleware.ExtractAndSetAcceptLanguage)
	r.Use(middleware.TraceId)

	routes := r.Group("/lawyer")
	// todo a lots of route....
	lawyerRoutes.RegisterUserApi(routes)
	//lawyerRoutes.RegisterOtherApi(routes)

	// plugin routes
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
func heartBeat(ctx *gin.Context) { ctx.String(200, "OK") }
