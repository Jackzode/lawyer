package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/commons/utils/checker"
	services "github.com/lawyer/initServer/initServices"
	"github.com/lawyer/middleware"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"net/url"
)

// UserController user controller
type UserController struct {
}

func NewUserController() *UserController {
	return &UserController{}
}

// GetUserInfoByUserID get user info, if user no login response http code is 200, but user info is null
func (uc *UserController) GetUserInfoByUserID(ctx *gin.Context) {
	token := middleware.ExtractToken(ctx)
	if len(token) == 0 {
		handler.HandleResponse(ctx, nil, nil)
		return
	}
	//从cache中获取userinfo， key是token， 再根据user id获取user status，最后写入缓存
	userCacheInfo, _ := services.AuthService.GetUserCacheInfo(ctx, token)
	if userCacheInfo == nil {
		handler.HandleResponse(ctx, nil, nil)
		return
	}
	//get user info from db
	userInfo, err := services.UserService.GetUserInfoByUserID(ctx, token, userCacheInfo.UserID)
	resp := &schema.GetCurrentLoginUserInfoResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.RoleID, err = services.UserRoleRelService.GetUserRole(ctx, userInfo.ID)
	if err != nil {
		log.Error(err)
	}
	resp.Avatar = services.SiteInfoCommonService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status)
	resp.AccessToken = token
	resp.HavePassword = len(userInfo.Pass) > 0
	//set cookie
	uc.setVisitCookies(ctx, userCacheInfo.VisitToken, false)
	handler.HandleResponse(ctx, err, resp)
}

// GetOtherUserInfoByUsername godoc
// @Summary GetOtherUserInfoByUsername
// @Description GetOtherUserInfoByUsername
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param username query string true "username"
// @Success 200 {object} handler.RespBody{data=schema.GetOtherUserInfoResp}
// @Router /answer/api/v1/personal/user/info [get]
func (uc *UserController) GetOtherUserInfoByUsername(ctx *gin.Context) {
	req := &schema.GetOtherUserInfoByUsernameReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := services.UserService.GetOtherUserInfoByUsername(ctx, req.Username)
	handler.HandleResponse(ctx, err, resp)
}

// UserEmailLogin godoc
// @Summary UserEmailLogin
// @Description UserEmailLogin
// @Tags User
// @Accept json
// @Produce json
// @Param data body schema.UserEmailLoginReq true "UserEmailLogin"
// @Success 200 {object} handler.RespBody{data=schema.UserLoginResp}
// @Router /answer/api/v1/user/login/email [post]
func (uc *UserController) UserEmailLogin(ctx *gin.Context) {
	req := &schema.UserEmailLoginReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := services.CaptchaService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionPassword, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	resp, err := services.UserService.EmailLogin(ctx, req)
	if err != nil {
		_, _ = services.CaptchaService.ActionRecordAdd(ctx, entity.CaptchaActionPassword, ctx.ClientIP())
		errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "e_mail",
			ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.EmailOrPasswordWrong),
		})
		handler.HandleResponse(ctx, errors.BadRequest(reason.EmailOrPasswordWrong), errFields)
		return
	}
	if !isAdmin {
		services.CaptchaService.ActionRecordDel(ctx, entity.CaptchaActionPassword, ctx.ClientIP())
	}
	uc.setVisitCookies(ctx, resp.VisitToken, true)
	handler.HandleResponse(ctx, nil, resp)
}

// RetrievePassWord godoc
// @Summary RetrievePassWord
// @Description RetrievePassWord
// @Tags User
// @Accept  json
// @Produce  json
// @Param data body schema.UserRetrievePassWordRequest  true "UserRetrievePassWordRequest"
// @Success 200 {string} string ""
// @Router /answer/api/v1/user/password/reset [post]
func (uc *UserController) RetrievePassWord(ctx *gin.Context) {
	req := &schema.UserRetrievePassWordRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := services.CaptchaService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}
	err := services.UserService.RetrievePassWord(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UseRePassWord godoc
// @Summary UseRePassWord
// @Description UseRePassWord
// @Tags User
// @Accept  json
// @Produce  json
// @Param data body schema.UserRePassWordRequest  true "UserRePassWordRequest"
// @Success 200 {string} string ""
// @Router /answer/api/v1/user/password/replacement [post]
func (uc *UserController) UseRePassWord(ctx *gin.Context) {
	req := &schema.UserRePassWordRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.Content = services.EmailService.VerifyUrlExpired(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.Forbidden(reason.EmailVerifyURLExpired),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeURLExpired})
		return
	}

	err := services.UserService.UpdatePasswordWhenForgot(ctx, req)
	services.CaptchaService.ActionRecordDel(ctx, entity.CaptchaActionPassword, ctx.ClientIP())
	handler.HandleResponse(ctx, err, nil)
}

// UserLogout user logout
// @Summary user logout
// @Description user logout
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/logout [get]
func (uc *UserController) UserLogout(ctx *gin.Context) {
	accessToken := middleware.ExtractToken(ctx)
	if len(accessToken) == 0 {
		handler.HandleResponse(ctx, nil, nil)
		return
	}
	_ = services.AuthService.RemoveUserCacheInfo(ctx, accessToken)
	_ = services.AuthService.RemoveAdminUserCacheInfo(ctx, accessToken)
	visitToken, _ := ctx.Cookie(constant.UserVisitCookiesCacheKey)
	_ = services.AuthService.RemoveUserVisitCacheInfo(ctx, visitToken)
	handler.HandleResponse(ctx, nil, nil)
}

// UserRegisterByEmail godoc
// @Summary UserRegisterByEmail
// @Description UserRegisterByEmail
// @Tags User
// @Accept json
// @Produce json
// @Param data body schema.UserRegisterReq true "UserRegisterReq"
// @Success 200 {object} handler.RespBody{data=schema.UserLoginResp}
// @Router /answer/api/v1/user/register/email [post]
func (uc *UserController) UserRegisterByEmail(ctx *gin.Context) {
	// check whether site allow register or not
	/*siteInfo, err := services.SiteInfoCommonService.GetSiteLogin(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !siteInfo.AllowNewRegistrations || !siteInfo.AllowEmailRegistrations {
		handler.HandleResponse(ctx, errors.BadRequest(reason.NotAllowedRegistration), nil)
		return
	}*/

	req := &schema.UserRegisterReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//同上
	/*if !checker.EmailInAllowEmailDomain(req.Email, siteInfo.AllowEmailDomains) {
		handler.HandleResponse(ctx, errors.BadRequest(reason.EmailIllegalDomainError), nil)
		return
	}*/
	req.IP = ctx.ClientIP()
	//就只是根据uuid获取用户身份，判断是否是管理员，
	//我尼玛新用户注册和管理员有鸡巴毛关系，这里的逻辑也可以删掉了
	/*isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := services.CaptchaService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, req.IP, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}*/

	resp, err := services.UserService.UserRegisterByEmail(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
	} else {
		handler.HandleResponse(ctx, err, resp)
	}
}

func (uc *UserController) UserVerifyEmail(ctx *gin.Context) {
	req := &schema.UserVerifyEmailReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//VerifyUrlExpired 根据code从缓存中获取content
	req.Content = services.EmailService.VerifyUrlExpired(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.Forbidden(reason.EmailVerifyURLExpired),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeURLExpired})
		return
	}
	//
	resp, err := services.UserService.UserVerifyEmail(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	services.CaptchaService.ActionRecordDel(ctx, entity.CaptchaActionEmail, ctx.ClientIP())
	handler.HandleResponse(ctx, err, resp)
}

// UserVerifyEmailSend godoc
// @Summary UserVerifyEmailSend
// @Description UserVerifyEmailSend
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param captcha_id query string false "captcha_id"  default()
// @Param captcha_code query string false "captcha_code"  default()
// @Success 200 {string} string ""
// @Router /answer/api/v1/user/email/verification/send [post]
func (uc *UserController) UserVerifyEmailSend(ctx *gin.Context) {
	req := &schema.UserVerifyEmailSendReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	userInfo := middleware.GetUserInfoFromContext(ctx)
	if userInfo == nil {
		handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
		return
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := services.CaptchaService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	err := services.UserService.UserVerifyEmailSend(ctx, userInfo.UserID)
	handler.HandleResponse(ctx, err, nil)
}

// UserModifyPassWord godoc
// @Summary UserModifyPassWord
// @Description UserModifyPassWord
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UserModifyPasswordReq  true "UserModifyPasswordReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/password [put]
func (uc *UserController) UserModifyPassWord(ctx *gin.Context) {
	req := &schema.UserModifyPasswordReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.AccessToken = middleware.ExtractToken(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := services.CaptchaService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionPassword, req.UserID,
			req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
		_, err := services.CaptchaService.ActionRecordAdd(ctx, entity.CaptchaActionPassword, req.UserID)
		if err != nil {
			log.Error(err)
		}
	}

	oldPassVerification, err := services.UserService.UserModifyPassWordVerification(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !oldPassVerification {
		errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "old_pass",
			ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.OldPasswordVerificationFailed),
		})
		handler.HandleResponse(ctx, errors.BadRequest(reason.OldPasswordVerificationFailed), errFields)
		return
	}

	if req.OldPass == req.Pass {
		errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "pass",
			ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.NewPasswordSameAsPreviousSetting),
		})
		handler.HandleResponse(ctx, errors.BadRequest(reason.NewPasswordSameAsPreviousSetting), errFields)
		return
	}
	err = services.UserService.UserModifyPassword(ctx, req)
	if err == nil {
		services.CaptchaService.ActionRecordDel(ctx, entity.CaptchaActionPassword, req.UserID)
	}
	handler.HandleResponse(ctx, err, nil)
}

// UserUpdateInfo update user info
// @Summary UserUpdateInfo update user info
// @Description UserUpdateInfo update user info
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "access-token"
// @Param data body schema.UpdateInfoRequest true "UpdateInfoRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/info [put]
func (uc *UserController) UserUpdateInfo(ctx *gin.Context) {
	req := &schema.UpdateInfoRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)
	errFields, err := services.UserService.UpdateInfo(ctx, req)
	for _, field := range errFields {
		field.ErrorMsg = translator.Tr(utils.GetLang(ctx), field.ErrorMsg)
	}
	handler.HandleResponse(ctx, err, errFields)
}

// UserUpdateInterface update user interface config
// @Summary UserUpdateInterface update user interface config
// @Description UserUpdateInterface update user interface config
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "access-token"
// @Param data body schema.UpdateUserInterfaceRequest true "UpdateInfoRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/interface [put]
func (uc *UserController) UserUpdateInterface(ctx *gin.Context) {
	req := &schema.UpdateUserInterfaceRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserId = middleware.GetLoginUserIDFromContext(ctx)
	err := services.UserService.UserUpdateInterface(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

func (uc *UserController) ActionRecord(ctx *gin.Context) {
	req := &schema.ActionRecordReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	//这里是从ctx里获取的user info
	userinfo := middleware.GetUserInfoFromContext(ctx)
	if userinfo != nil {
		req.UserID = userinfo.UserID
	}
	req.IP = ctx.ClientIP()
	resp := &schema.ActionRecordResp{}
	//role id 是2和3，就是管理员， 管理员不需要验证
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if isAdmin {
		resp.Verify = false
		handler.HandleResponse(ctx, nil, resp)
	} else {
		var err error
		unit := services.CaptchaService.GetActionRecordUnit(ctx, req)
		verificationResult := services.CaptchaService.ValidationStrategy(ctx, unit, req.Action)
		if !verificationResult {
			resp.CaptchaID, resp.CaptchaImg, err = services.CaptchaService.GenerateCaptcha(ctx)
			resp.Verify = true
		}
		handler.HandleResponse(ctx, err, resp)
	}

}

func (uc *UserController) UserRegisterCaptcha(ctx *gin.Context) {
	resp := &schema.ActionRecordResp{}
	CaptchaID, CaptchaImg, err := utils.GenerateCaptcha(ctx)
	resp.Verify = true
	resp.CaptchaID = CaptchaID
	resp.CaptchaImg = CaptchaImg
	handler.HandleResponse(ctx, err, resp)
}

// GetUserNotificationConfig get user's notification config
// @Summary get user's notification config
// @Description get user's notification config
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=schema.GetUserNotificationConfigResp}
// @Router /answer/api/v1/user/notification/config [post]
func (uc *UserController) GetUserNotificationConfig(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)
	resp, err := services.UserNotificationConfigService.GetUserNotificationConfig(ctx, userID)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateUserNotificationConfig update user's notification config
// @Summary update user's notification config
// @Description update user's notification config
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateUserNotificationConfigReq true "UpdateUserNotificationConfigReq"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/user/notification/config [put]
func (uc *UserController) UpdateUserNotificationConfig(ctx *gin.Context) {
	req := &schema.UpdateUserNotificationConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	err := services.UserNotificationConfigService.UpdateUserNotificationConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UserChangeEmailSendCode send email to the user email then change their email
// @Summary send email to the user email then change their email
// @Description send email to the user email then change their email
// @Tags User
// @Accept json
// @Produce json
// @Param data body schema.UserChangeEmailSendCodeReq true "UserChangeEmailSendCodeReq"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/user/email/change/code [post]
func (uc *UserController) UserChangeEmailSendCode(ctx *gin.Context) {
	req := &schema.UserChangeEmailSendCodeReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	// If the user is not logged in, the api cannot be used.
	// If the user email is not verified, that also can use this api to modify the email.
	if len(req.UserID) == 0 {
		handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
		return
	}
	// check whether email allow register or not
	siteInfo, err := services.SiteInfoCommonService.GetSiteLogin(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !checker.EmailInAllowEmailDomain(req.Email, siteInfo.AllowEmailDomains) {
		handler.HandleResponse(ctx, errors.BadRequest(reason.EmailIllegalDomainError), nil)
		return
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)

	if !isAdmin {
		captchaPass := services.CaptchaService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEditUserinfo, req.UserID, req.CaptchaID, req.CaptchaCode)
		services.CaptchaService.ActionRecordAdd(ctx, entity.CaptchaActionEditUserinfo, req.UserID)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	resp, err := services.UserService.UserChangeEmailSendCode(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, resp)
		return
	}
	if !isAdmin {
		services.CaptchaService.ActionRecordDel(ctx, entity.CaptchaActionEditUserinfo, ctx.ClientIP())
	}

	handler.HandleResponse(ctx, err, nil)
}

// UserChangeEmailVerify user change email verification
// @Summary user change email verification
// @Description user change email verification
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UserChangeEmailVerifyReq true "UserChangeEmailVerifyReq"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/user/email [put]
func (uc *UserController) UserChangeEmailVerify(ctx *gin.Context) {
	req := &schema.UserChangeEmailVerifyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.Content = services.EmailService.VerifyUrlExpired(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.Forbidden(reason.EmailVerifyURLExpired),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeURLExpired})
		return
	}

	resp, err := services.UserService.UserChangeEmailVerify(ctx, req.Content)
	services.CaptchaService.ActionRecordDel(ctx, entity.CaptchaActionEmail, ctx.ClientIP())
	handler.HandleResponse(ctx, err, resp)
}

// UserRanking get user ranking
// @Summary get user ranking
// @Description get user ranking
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=schema.UserRankingResp}
// @Router /answer/api/v1/user/ranking [get]
func (uc *UserController) UserRanking(ctx *gin.Context) {
	resp, err := services.UserService.UserRanking(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// UserUnsubscribeNotification unsubscribe notification
// @Summary unsubscribe notification
// @Description unsubscribe notification
// @Tags User
// @Accept json
// @Produce json
// @Param data body schema.UserUnsubscribeNotificationReq true "UserUnsubscribeNotificationReq"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/user/notification/unsubscribe [put]
func (uc *UserController) UserUnsubscribeNotification(ctx *gin.Context) {
	req := &schema.UserUnsubscribeNotificationReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.Content = services.EmailService.VerifyUrlExpired(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.Forbidden(reason.EmailVerifyURLExpired),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeURLExpired})
		return
	}

	err := services.UserService.UserUnsubscribeNotification(ctx, req)
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
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	resp, err := services.UserService.SearchUserListByName(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

func (uc *UserController) setVisitCookies(ctx *gin.Context, visitToken string, force bool) {

	cookie, err := ctx.Cookie(constant.UserVisitCookiesCacheKey)
	if err == nil && len(cookie) > 0 && !force {
		return
	}
	general, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site general error: %v", err)
		return
	}
	parsedURL, err := url.Parse(general.SiteUrl)
	if err != nil {
		log.Errorf("parse url error: %v", err)
		return
	}
	ctx.SetCookie(constant.UserVisitCookiesCacheKey,
		visitToken, constant.UserVisitCacheTime, "/", parsedURL.Host, true, true)
}