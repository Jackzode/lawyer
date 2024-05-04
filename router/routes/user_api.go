package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
	"github.com/lawyer/controller_admin"
	"github.com/lawyer/middleware"
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
	rg.GET("/logout", c.UserLogout) //本地把token删除就行了，其实服务端不需要干啥

	rg.PUT("/update/info", c.UserUpdateInfo)                                   //need login
	rg.GET("/getUserInfo", c.GetUserInfoByUserID)                              //need login
	rg.GET("/action/record", c.ActionRecord)                                   //need login
	rg.PUT("/change/password", middleware.AccessToken(), c.UserModifyPassWord) //need login
	rg.POST("/email/verification/send", c.UserVerifyEmailSend)                 //need login
	rg.POST("/email/change/code", c.UserChangeEmailSendCode)                   //need login
	rg.GET("/info/search", c.SearchUserListByName)                             //need login
	rg.PUT("/change/email", c.UserChangeEmailVerify)
	/*
		忘记密码的逻辑：用户先填写邮箱，点击忘记密码进入/password/reset接口，
		然后邮箱会收到一个链接，链接带一个code，点开链接是一个页面，上下两个
		输入框，新密码，确认密码，点击提交后走到/password/replacement接口里，完成重置密码
		两个输入框是否一致在前端判断，传给接口只是一个新的password。
		目前看来缺少一个接口（页面），就是邮件里的链接，点击链接展现两个输入框，一个按钮。这个
		前端就可以完成了，不需要后端接口，前端加一个页面就可以
	*/
	rg.POST("/password/reset", c.RetrievePassWord)
	rg.POST("/password/replacement", c.UserReplacePassWord)

	rg.GET("/personal/info", c.GetOtherUserInfoByUsername)
	rg.PUT("/interface/lang", c.UserUpdateInterfaceLang)

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
