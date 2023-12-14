package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"lawyer/common"
	"lawyer/service"
	"lawyer/types"
	"lawyer/utils"
)

type UserController struct {
	//暂时先不注入，随后再说
}

func (u *UserController) GetUserInfo(ctx *gin.Context) {

}

func (u *UserController) UserLoginByEmail(ctx *gin.Context) {
	loginReq := &types.UserEmailLoginReq{}
	err := ctx.ShouldBind(loginReq)
	if err != nil {
		ResponseHandler(ctx, common.ParamErrCode, common.RequestParamErrMsg, nil)
		return
	}
	userService := new(service.UserService)
	userInfo, err := userService.EmailLogin(ctx, loginReq)
	if err != nil {
		log.Log().Msg(err.Error())
		return
	}
	//generate token
	token, err := utils.CreateToken(userInfo.Username, userInfo.ID, userInfo.RoleID)
	userService.SaveUserToken(ctx, userInfo.ID, token)
	if err != nil {
		fmt.Println("get token err: ", err)
		return
	}
	//response
	data := types.LoginResponse{
		UserInfo: *userInfo,
		Token:    token,
	}
	ResponseHandler(ctx, common.OK, common.Success, data)
}

func (u *UserController) RegisterByEmail(ctx *gin.Context) {
	req := types.UserRegisterReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		fmt.Printf("param err : %s", err.Error())
		ResponseHandler(ctx, common.ParamErrCode, common.RequestParamErrMsg, nil)
		return
	}
	//1. get email from mysql, if exist , return
	userService := new(service.UserService)
	userInfo, get, err := userService.GetUserInfoByEmail(req)
	if err != nil {
		ResponseHandler(ctx, common.EmailRegisteredCode, err.Error(), nil)
		return
	}
	if get {
		if userInfo.Status == 0 {
			ResponseHandler(ctx, common.EmailRegisteredCode, common.EmailRegisteredMsg, nil)
			return
		}
		if userInfo.Status == 1 {
			ResponseHandler(ctx, common.UserAccountSuspended, common.UserAccountSuspendedMsg, nil)
			return
		}
		if userInfo.Status == 2 {
			ResponseHandler(ctx, common.UserAccountException, common.UserAccountExceptionMsg, nil)
			return
		}
	}
	//2. check captcha
	captcha := userService.GetCaptchaCode(ctx)
	if req.Captcha != captcha {
		ResponseHandler(ctx, common.CaptchaErrCode, common.CaptchaErrMsg, nil)
		return
	}
	//3. save on db
	newUserInfo, err := userService.SaveUserInfo(req)
	if err != nil {
		ResponseHandler(ctx, common.InternalErrorCode, err.Error(), nil)
		return
	}
	//response return userInfo
	ResponseHandler(ctx, common.OK, common.RegisterSuccessMsg, newUserInfo)
}

// GetCaptchaByEmail 发送验证码，并将验证码保存到redis里，
// 当用户提交表单数据时，对比邮箱验证码
func (u *UserController) GetCaptchaByEmail(ctx *gin.Context) {
	email := ctx.PostForm("email")
	if email == "" {
		fmt.Println("email is invalid")
		ResponseHandler(ctx, common.ParamErrCode, common.RequestParamErrMsg, nil)
		return
	}
	s := &service.UserService{}
	err := s.SendCaptchaByEmail(ctx, email)
	if err != nil {
		ResponseHandler(ctx, common.InternalErrorCode, err.Error(), nil)
		return
	}
	ResponseHandler(ctx, common.OK, common.Success, nil)
}

func (u *UserController) UserLogout(context *gin.Context) {

	//delete token in redis
	s := &service.UserService{}
	key := ""
	s.DeleteUserToken(context, key)
	ResponseHandler(context, common.OK, common.Success, nil)
}
