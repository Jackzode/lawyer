package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	glog "github.com/lawyer/commons/logger"
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

	//register common middleware
	r.Use(middleware.ExtractAndSetLanguage())
	r.Use(middleware.TraceId())
	r.Use(middleware.RecoverPanic())

	router := r.Group("/lawyer")

	routes.RegisterUserApi(router)

	//routes.RegisterQuestionApi(router)
	//
	//routes.RegisterOtherApi(router)
	//
	//routes.RegisterAnswerApi(router)
	//
	//routes.RegisterCommentApi(router)
	//
	//routes.RegisterNotificationApi(router)
	//
	//routes.RegisterReportApi(router)
	//
	//routes.RegisterTagApi(router)
	//
	//routes.RegisterRevisionApi(router)
	//管理员用的后台接口
	//routes.RegisterAdminUserApi(router)

}

func heartBeats(ctx *gin.Context) {
	fmt.Println("heart beat ....")
	glog.Klog.Info("heart beat ...")
	ctx.String(200, "OK")
}

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
