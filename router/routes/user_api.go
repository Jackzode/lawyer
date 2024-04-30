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

func RegisterUserApi(r *gin.RouterGroup) {

	c := controller.NewUserController()
	ac := controller_admin.NewUserAdminController()
	r.GET("/user/info", c.GetUserInfoByUserID) //need login
	r.GET("/user/action/record", c.ActionRecord)
	//skip plugin
	routerGroup := r.Group("", middleware.BanAPIForUserCenter)

	/*
		获取验证码的接口，返回给端上id， 验证码图片，把id和答案保存在redis里
	*/
	routerGroup.GET("/user/register/captcha", c.UserRegisterCaptcha)
	/*

	 */
	routerGroup.POST("/user/register/email", c.UserRegisterByEmail)
	/*

	 */
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
