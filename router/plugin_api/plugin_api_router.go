package plugin_api

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
	"github.com/lawyer/controller_admin"
)

type PluginAPIRouter struct {
	connectorController  *controller.ConnectorController
	userCenterController *controller.UserCenterController
}

func NewPluginAPIRouter(
	connectorController *controller.ConnectorController,
	userCenterController *controller.UserCenterController,
) *PluginAPIRouter {
	return &PluginAPIRouter{
		connectorController:  connectorController,
		userCenterController: userCenterController,
	}
}

func (pr *PluginAPIRouter) RegisterUnAuthConnectorRouter(r *gin.RouterGroup) {
	// connector plugin
	connectorController := pr.connectorController
	r.GET(controller.ConnectorLoginRouterPrefix+":name", connectorController.ConnectorLoginDispatcher)
	r.GET(controller.ConnectorRedirectRouterPrefix+":name", connectorController.ConnectorRedirectDispatcher)
	r.GET("/connector/info", connectorController.ConnectorsInfo)
	r.POST("/connector/binding/email", connectorController.ExternalLoginBindingUserSendEmail)

	// user center plugin
	//登录时调用
	r.GET("/user-center/agent", pr.userCenterController.UserCenterAgent)
	r.GET("/user-center/personal/branding", pr.userCenterController.UserCenterPersonalBranding)
	r.GET(controller.UserCenterLoginRouter, pr.userCenterController.UserCenterLoginRedirect)
	r.GET(controller.UserCenterSignUpRedirectRouter, pr.userCenterController.UserCenterSignUpRedirect)
	r.GET("/user-center/login/callback", pr.userCenterController.UserCenterLoginCallback)
	r.GET("/user-center/sign-up/callback", pr.userCenterController.UserCenterSignUpCallback)
}

func (pr *PluginAPIRouter) RegisterAuthUserConnectorRouter(r *gin.RouterGroup) {
	connectorController := pr.connectorController
	r.GET("/connector/user/info", connectorController.ConnectorsUserInfo)
	r.DELETE("/connector/user/unbinding", connectorController.ExternalLoginUnbinding)

	r.GET("/user-center/user/settings", pr.userCenterController.UserCenterUserSettings)
}

func (pr *PluginAPIRouter) RegisterAuthAdminConnectorRouter(r *gin.RouterGroup) {
	r.GET("/user-center/agent", pr.userCenterController.UserCenterAdminFunctionAgent)
}

func RegisterPluginApi(r *gin.RouterGroup) {
	// plugin
	c := controller_admin.NewPluginController()
	r.GET("/plugin/status", c.GetAllPluginStatus)
	r.GET("/plugins", c.GetPluginList)
	r.PUT("/plugin/status", c.UpdatePluginStatus)
	r.GET("/plugin/config", c.GetPluginConfig)
	r.PUT("/plugin/config", c.UpdatePluginConfig)
}
