package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/service"
)

// UserController user controller, no need login
type UserController struct {
}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) UserEmailLogin(ctx *gin.Context) {
	req := &schema.UserEmailLoginReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//验证码是否正确
	captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionPassword, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
	if !captchaPass {
		handler.HandleResponse(ctx, errors.New(reason.CaptchaVerificationFailed), nil)
		return
	}
	//查询db，生成userCacheInfo,这里包含了role id
	resp, err := service.UserServicer.EmailLogin(ctx, req)
	if err != nil {
		//记录登陆失败的次数，先注释掉，和主业务没关系的统统注释掉
		//service.CaptchaServicer.ActionRecordAdd(ctx, entity.CaptchaActionPassword, ctx.ClientIP())
		handler.HandleResponse(ctx, errors.New(reason.EmailOrPasswordWrong), nil)
		return
	}

	//service.CaptchaServicer.ActionRecordDel(ctx, entity.CaptchaActionPassword, ctx.ClientIP())

	//uc.setVisitCookies(ctx, resp.VisitToken, true)
	handler.HandleResponse(ctx, nil, resp)
}

/*

 */
// GetUserInfoByUserID get user info, if user no login response http code is 200, but user info is null
func (uc *UserController) GetUserInfoByUserID(ctx *gin.Context) {
	//token := utils.ExtractToken(ctx)
	//if len(token) == 0 {
	//	handler.HandleResponse(ctx, nil, nil)
	//	return
	//}
	//从cache中获取userinfo， key是token
	//userCacheInfo, _ := service.AuthServicer.GetUserCacheInfo(ctx, token)
	//if userCacheInfo == nil {
	//	handler.HandleResponse(ctx, nil, nil)
	//	return
	//}
	//get user info from db
	uid := utils.GetUidFromTokenByCtx(ctx)
	userInfo, err := service.UserServicer.GetUserInfoByUserID(ctx, uid)
	resp := &schema.GetCurrentLoginUserInfoResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.RoleID = 1
	//resp.RoleID, err = service.UserRoleRelServicer.GetUserRole(ctx, userInfo.ID)
	//if err != nil {
	//	glog.Slog.Error(err)
	//}
	//拼接头像, todo
	resp.Avatar = service.SiteInfoCommonServicer.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status)
	//resp.AccessToken = token
	resp.HavePassword = len(userInfo.Pass) > 0
	//set cookie
	//uc.setVisitCookies(ctx, userCacheInfo.VisitToken, false)
	handler.HandleResponse(ctx, err, resp)
}

// 根据用户名获取用户信息，不需要登录
func (uc *UserController) GetOtherUserInfoByUsername(ctx *gin.Context) {
	req := &schema.GetOtherUserInfoByUsernameReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//核心逻辑
	resp, err := service.UserServicer.GetOtherUserInfoByUsername(ctx, req.Username)
	handler.HandleResponse(ctx, err, resp)
}

// 找回密码
func (uc *UserController) RetrievePassWord(ctx *gin.Context) {
	req := &schema.UserRetrievePassWordRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//校对验证码
	captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
	fmt.Println("captchaPass: ", captchaPass)
	if !captchaPass {
		handler.HandleResponse(ctx, errors.New(reason.CaptchaVerificationFailed), nil)
		return
	}
	//core logic
	err := service.UserServicer.RetrievePassWord(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// @Router /answer/api/v1/user/password/replacement [post]
func (uc *UserController) UserReplacePassWord(ctx *gin.Context) {
	req := &schema.UserRePassWordRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//这个code是/password/reset接口生成的，里面存的是email和uid
	req.Content = service.EmailServicer.VerifyEmailByCode(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.New(reason.EmailVerifyURLExpired), nil)
		return
	}
	//更新db中的密码
	err := service.UserServicer.UpdatePasswordWhenForgot(ctx, req)
	//service.CaptchaServicer.ActionRecordDel(ctx, entity.CaptchaActionPassword, ctx.ClientIP())
	handler.HandleResponse(ctx, err, nil)
}

/*
退出登录，直接删除用户token就行
*/
// @Router /answer/api/v1/user/logout [get]
func (uc *UserController) UserLogout(ctx *gin.Context) {
	accessToken := utils.ExtractToken(ctx)
	if len(accessToken) == 0 {
		handler.HandleResponse(ctx, nil, nil)
		return
	}
	_ = service.AuthServicer.RemoveUserCacheInfo(ctx, accessToken)
	//_ = service.AuthServicer.RemoveAdminUserCacheInfo(ctx, accessToken)
	//visitToken, _ := ctx.Cookie(constant.UserVisitCookiesCacheKey)
	//_ = service.AuthServicer.RemoveUserVisitCacheInfo(ctx, visitToken)
	handler.HandleResponse(ctx, nil, nil)
}

// @Router /lawyer/user/register/email [post]
func (uc *UserController) UserRegisterByEmail(ctx *gin.Context) {
	req := &schema.UserRegisterReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.IP = ctx.ClientIP()
	//对比验证码是否正确
	captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, req.IP, req.CaptchaID, req.CaptchaCode)
	if !captchaPass {
		handler.HandleResponse(ctx, errors.New(reason.CaptchaVerificationFailed), nil)
		return
	}
	//核心逻辑
	resp, err := service.UserServicer.UserRegisterByEmail(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

func (uc *UserController) UserVerifyEmail(ctx *gin.Context) {
	req := &schema.UserVerifyEmailReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//VerifyEmailByCode 根据code从缓存中获取content,包含email和uid
	req.Content = service.EmailServicer.VerifyEmailByCode(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.New(reason.EmailVerifyURLExpired), nil)
		return
	}
	//验证邮箱
	resp, err := service.UserServicer.UserVerifyEmail(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	service.CaptchaServicer.ActionRecordDel(ctx, entity.CaptchaActionEmail, ctx.ClientIP())
	handler.HandleResponse(ctx, err, resp)
}

// @Router /answer/api/v1/user/email/verification/send [post]
func (uc *UserController) UserVerifyEmailSend(ctx *gin.Context) {
	req := &schema.UserVerifyEmailSendReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	uid := utils.GetUidFromTokenByCtx(ctx)
	captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
	if !captchaPass {
		handler.HandleResponse(ctx, errors.New(reason.CaptchaVerificationFailed), nil)
		return
	}

	err := service.UserServicer.UserVerifyEmailSend(ctx, uid)
	handler.HandleResponse(ctx, err, nil)
}

// @Router /answer/api/v1/user/password [put]
func (uc *UserController) UserModifyPassWord(ctx *gin.Context) {
	req := &schema.UserModifyPasswordReq{}
	fmt.Println("req ", req)
	if handler.BindAndCheck(ctx, req) {
		return
	}
	uid := utils.GetUidFromTokenByCtx(ctx)
	req.UserID = uid
	//校对验证码
	captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionPassword, req.UserID,
		req.CaptchaID, req.CaptchaCode)
	if !captchaPass {
		handler.HandleResponse(ctx, errors.New(reason.CaptchaVerificationFailed), nil)
		return
	}
	//记录action
	service.CaptchaServicer.ActionRecordAdd(ctx, entity.CaptchaActionPassword, req.UserID)

	//验证用户老密码是否正确
	oldPassVerification := service.UserServicer.UserPassWordVerification(ctx, req.UserID, req.OldPass)
	if !oldPassVerification {
		handler.HandleResponse(ctx, errors.New(reason.OldPasswordVerificationFailed), nil)
		return
	}

	//修改密码时新密码和老密码不能一样
	if req.OldPass == req.Pass {
		handler.HandleResponse(ctx, errors.New(reason.NewPasswordSameAsPreviousSetting), nil)
		return
	}
	err := service.UserServicer.UserModifyPassword(ctx, req)
	if err == nil {
		//删除这个action记录
		service.CaptchaServicer.ActionRecordDel(ctx, entity.CaptchaActionPassword, req.UserID)
	}
	handler.HandleResponse(ctx, err, nil)
}

// @Router /answer/api/v1/user/info [put]
func (uc *UserController) UserUpdateInfo(ctx *gin.Context) {
	req := &schema.UpdateInfoRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//从token里获取用户信息
	req.UserID = utils.GetUidFromTokenByCtx(ctx)
	err := service.UserServicer.UpdateInfo(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// @Router /answer/api/v1/user/interface [put]
func (uc *UserController) UserUpdateInterfaceLang(ctx *gin.Context) {
	req := &schema.UpdateUserInterfaceRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//req.UserId = middleware.GetLoginUserIDFromContext(ctx)
	req.UserId = utils.GetUidFromTokenByCtx(ctx)
	if !translator.CheckLanguageIsValid(req.Language) {
		handler.HandleResponse(ctx, errors.New(reason.LangNotFound), nil)
		return
	}
	//根据uid更新用户user表的language字段
	err := service.UserServicer.UserUpdateInterface(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// 判断当前action是否需要验证码
func (uc *UserController) ActionRecord(ctx *gin.Context) {
	req := &schema.ActionRecordReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//这里是从ctx里获取的user info
	//userinfo := middleware.GetUserInfoFromContext(ctx)
	//if userinfo != nil {
	//	req.UserID = userinfo.UserID
	//}
	uid := utils.GetUidFromTokenByCtx(ctx)
	req.UserID = uid
	req.IP = ctx.ClientIP()
	resp := &schema.ActionRecordResp{}
	//role id 是2和3，就是管理员， 管理员不需要验证
	//service.CaptchaServicer.ActionRecordAdd(ctx, req.Action, req.IP)
	unit := service.CaptchaServicer.GetActionRecordUnit(ctx, req)
	//对于当前action是否需要验证码
	verificationResult := service.CaptchaServicer.ValidationStrategy(ctx, unit, req.Action)
	//需要验证码
	var err error
	if verificationResult {
		resp.CaptchaID, resp.CaptchaImg, err = service.CaptchaServicer.GenerateCaptcha(ctx)
		resp.Verify = true
	}
	handler.HandleResponse(ctx, err, resp)

}

// 生成验证码，返回给端上
func (uc *UserController) UserRegisterCaptcha(ctx *gin.Context) {
	resp := &schema.ActionRecordResp{}
	key, base64, err := service.CaptchaServicer.GenerateCaptcha(ctx)
	resp.Verify = true
	resp.CaptchaID = key
	resp.CaptchaImg = base64
	handler.HandleResponse(ctx, err, resp)
}

// todo @Router /answer/api/v1/user/notification/config [post]
func (uc *UserController) GetUserNotificationConfig(ctx *gin.Context) {
	//userID := middleware.GetLoginUserIDFromContext(ctx)
	userID := utils.GetUidFromTokenByCtx(ctx)
	resp, err := service.UserNotificationConfigService.GetUserNotificationConfig(ctx, userID)
	handler.HandleResponse(ctx, err, resp)
}

// todo @Router /answer/api/v1/user/notification/config [put]
func (uc *UserController) UpdateUserNotificationConfig(ctx *gin.Context) {
	req := &schema.UpdateUserNotificationConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	//req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.UserID = utils.GetUidFromTokenByCtx(ctx)
	err := service.UserNotificationConfigService.UpdateUserNotificationConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// @Router /answer/api/v1/user/email/change/code [post]
func (uc *UserController) UserChangeEmailSendCode(ctx *gin.Context) {
	req := &schema.UserChangeEmailSendCodeReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = utils.GetUidFromTokenByCtx(ctx)
	// If the user is not logged in, the api cannot be used.
	// If the user email is not verified, that also can use this api to modify the email.
	if len(req.UserID) == 0 {
		handler.HandleResponse(ctx, errors.New(reason.UnauthorizedError), nil)
		return
	}

	//校对验证码
	captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEditUserinfo, req.UserID, req.CaptchaID, req.CaptchaCode)
	//记录本次修改用户信息的操作
	service.CaptchaServicer.ActionRecordAdd(ctx, entity.CaptchaActionEditUserinfo, req.UserID)
	if !captchaPass {
		handler.HandleResponse(ctx, errors.New(reason.CaptchaVerificationFailed), nil)
		return
	}
	//核心逻辑
	err := service.UserServicer.UserChangeEmailSendCode(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	//删除这个操作记录
	service.CaptchaServicer.ActionRecordDel(ctx, entity.CaptchaActionEditUserinfo, ctx.ClientIP())
	handler.HandleResponse(ctx, err, nil)
}

// @Router /answer/api/v1/user/email [put]
func (uc *UserController) UserChangeEmailVerify(ctx *gin.Context) {
	req := &schema.UserChangeEmailVerifyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.Content = service.EmailServicer.VerifyEmailByCode(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.New(reason.EmailVerifyURLExpired), nil)
		return
	}
	//核心逻辑
	resp, err := service.UserServicer.UserChangeEmailVerify(ctx, req.Content)
	service.CaptchaServicer.ActionRecordDel(ctx, entity.CaptchaActionEmail, ctx.ClientIP())
	handler.HandleResponse(ctx, err, resp)
}

// @Router /answer/api/v1/user/ranking [get]
func (uc *UserController) UserRanking(ctx *gin.Context) {
	resp, err := service.UserServicer.UserRanking(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// @Router /answer/api/v1/user/notification/unsubscribe [put]
func (uc *UserController) UserUnsubscribeNotification(ctx *gin.Context) {
	req := &schema.UserUnsubscribeNotificationReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.Content = service.EmailServicer.VerifyEmailByCode(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.New(reason.EmailVerifyURLExpired), nil)
		return
	}

	err := service.UserServicer.UserUnsubscribeNotification(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// SearchUserListByName godoc
// @Summary SearchUserListByName
// @Description SearchUserListByName
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param username query string true "username"
// @Success 200 {object} handler.RespBody{data=schema.GetOtherUserInfoResp}
// @Router /answer/api/v1/user/info/search [get]
func (uc *UserController) SearchUserListByName(ctx *gin.Context) {
	req := &schema.GetOtherUserInfoByUsernameReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//根据token获取uid，我觉得不需要登录
	//req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	//req.UserID = utils.GetUidFromTokenByCtx(ctx)
	resp, err := service.UserServicer.SearchUserListByName(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

func (uc *UserController) setVisitCookies(ctx *gin.Context, visitToken string, force bool) {

	cookie, err := ctx.Cookie(constant.UserVisitCookiesCacheKey)
	if err == nil && len(cookie) > 0 && !force {
		return
	}
	//general, err := service.SiteInfoCommonServicer.GetSiteGeneral(ctx)
	//if err != nil {
	//	log.Errorf("get site general error: %v", err)
	//	return
	//}
	//parsedURL, err := url.Parse(general.SiteUrl)
	//if err != nil {
	//	log.Errorf("parse url error: %v", err)
	//	return
	//}
	//ctx.SetCookie(constant.UserVisitCookiesCacheKey, visitToken, constant.UserVisitCacheTime, "/", parsedURL.Host, true, true)
}
