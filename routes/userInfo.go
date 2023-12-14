package routes

import (
	"github.com/gin-gonic/gin"
	"lawyer/controller"
)

func InitUserApiRouter(engine *gin.RouterGroup) {
	engine.GET("/getUserInfo", (&controller.UserController{}).GetUserInfo)
	engine.GET("/loginByEmail", (&controller.UserController{}).UserLoginByEmail)
	engine.GET("/logout", (&controller.UserController{}).UserLogout)
	engine.POST("/registerByEmail", (&controller.UserController{}).RegisterByEmail)
	engine.POST("/getCaptcha", (&controller.UserController{}).GetCaptchaByEmail)
}
