package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
	"github.com/lawyer/controller_admin"
)

// todo
func RegisterUserApi(r *gin.RouterGroup) {
	c := controller.NewUserController()
	rg := r.Group("/user")
	/*获取验证码的接口，返回给端上id， 验证码图片，把id和答案保存在redis里*/
	rg.POST("/register/email", c.UserRegisterByEmail)
	rg.GET("/register/captcha", c.UserRegisterCaptcha)
	rg.POST("/email/verification", c.UserVerifyEmail)
	rg.POST("/login/email", c.UserEmailLogin)
	rg.PUT("/change/email", c.UserChangeEmailVerify)
	rg.POST("/password/reset", c.RetrievePassWord)
	rg.POST("/password/replacement", c.UserReplacePassWord)
	rg.GET("/logout", c.UserLogout)
	rg.GET("/personal/info", c.GetOtherUserInfoByUsername)
	rg.PUT("/interface/lang", c.UserUpdateInterfaceLang)
	rg.GET("/getUserInfo", c.GetUserInfoByUserID)              //need login
	rg.GET("/action/record", c.ActionRecord)                   //need login
	rg.PUT("/change/password", c.UserModifyPassWord)           //need login
	rg.POST("/email/verification/send", c.UserVerifyEmailSend) //need login
	rg.POST("/email/change/code", c.UserChangeEmailSendCode)   //need login
	rg.PUT("/update/info", c.UserUpdateInfo)                   //need login
	rg.GET("/info/search", c.SearchUserListByName)             //need login

	//todo
	// user
	//rg.POST("/user/email/change/code", middleware.BanAPIForUserCenter, c.UserChangeEmailSendCode)//need login
	//rg.POST("/user/email/verification/send", middleware.BanAPIForUserCenter, c.UserVerifyEmailSend)
	//rg.PUT("/user/password", middleware.BanAPIForUserCenter, c.UserModifyPassWord)
	//r.GET("/user/action/record", authUserMiddleware.Auth(), c.ActionRecord)

	//还没有看的
	rg.GET("/ranking", c.UserRanking)
	rg.PUT("/notification/unsubscribe", c.UserUnsubscribeNotification)
	rg.GET("/notification/config", c.GetUserNotificationConfig)    //need login
	rg.PUT("/notification/config", c.UpdateUserNotificationConfig) //need login

}

func RegisterAdminUserApi(r *gin.RouterGroup) {
	//管理员使用的后台接口
	ac := controller_admin.NewUserAdminController()
	r.GET("/users/page", ac.GetUserPage)
	r.PUT("/user/status", ac.UpdateUserStatus)
	r.PUT("/user/role", ac.UpdateUserRole)
	r.GET("/user/activation", ac.GetUserActivation)
	r.POST("/user/activation", ac.SendUserActivation)
	r.POST("/user", ac.AddUser)
	r.POST("/users", ac.AddUsers)
	r.PUT("/user/password", ac.UpdateUserPassword)
}
