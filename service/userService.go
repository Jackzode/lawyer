package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gopkg.in/gomail.v2"
	"lawyer/common"
	"lawyer/config"
	"lawyer/dao"
	"lawyer/dao/downstream"
	"lawyer/types"
	"lawyer/utils"
	"time"
)

type UserService struct {
}

func (u *UserService) GetCaptchaCode(ctx *gin.Context) string {
	key := common.CaptchaPrefix + ctx.ClientIP()
	captcha := dao.GetUserCaptcha(ctx, key)
	return captcha
}

func (u *UserService) GetUserInfoByEmail(req types.UserRegisterReq) (*types.UserInfo, bool, error) {
	userInfo := &types.UserInfo{}
	get, err := dao.FindUserByOneCondition(userInfo, "e_mail", req.Email)
	return userInfo, get, err
}

func (u *UserService) SaveUserInfo(req types.UserRegisterReq) (*types.UserInfo, error) {
	userInfo := &types.UserInfo{}
	userInfo.Status = 0
	userInfo.Username = req.Username
	userInfo.EMail = req.Email
	userInfo.RoleID = 1
	userInfo.PassWord = utils.BcryptHash(req.Password)
	userInfo.Bio = "I'm lazy, so I didn't write anything"
	userInfo.AuthorityGroup = 1
	userInfo.CreatedAt = time.Now().Unix()
	userInfo.HavePassword = true
	userInfo.LastLoginDate = time.Now().Unix()
	userInfo.Mobile = ""
	userInfo.Location = "earth"
	err := dao.SaveUserEmail(userInfo)
	return userInfo, err
}

func (u *UserService) SendCaptchaByEmail(ctx *gin.Context, email string) (err error) {
	recipient := email
	// 设置邮件内容
	subject := "lawyer verification code"
	//6 digits code
	code := utils.CreateCaptcha(6)
	body := fmt.Sprintf("code is : %s, \n Expires in one minute", code)
	// 创建邮件消息
	go func() {
		m := gomail.NewMessage()
		m.SetHeader("From", config.SmtpUsername)
		m.SetHeader("To", recipient)
		m.SetHeader("Subject", subject)
		m.SetBody("text/plain", body)
		// 创建邮件发送器
		d := gomail.NewDialer(config.SMTPHost, config.SmtpPort, config.SmtpUsername, config.SmtpPassWord)
		err = d.DialAndSend(m)
		fmt.Println("DialAndSend err=", err)
	}()
	//redis set key=uid value=code 60s
	key := "user_" + ctx.ClientIP()
	ex := downstream.RedisClient.SetEX(ctx, key, code, time.Minute*3)
	err = ex.Err()
	return
}

func (u *UserService) EmailLogin(ctx *gin.Context, req *types.UserEmailLoginReq) (*types.UserInfo, error) {

	//VerifyCaptcha
	//key := req.CaptchaID //key是固定的
	//captchaID := req.CaptchaCode
	//realCaptchaID := downstream.RedisClient.Get(ctx, key)
	//if captchaID != realCaptchaID.String() {
	//	err := errors.New("verification code not right")
	//	return nil, err
	//}
	//get userinfo
	userInfo := &types.UserInfo{}
	get, err := dao.FindUserByOneCondition(userInfo, "e_mail", req.Email)
	if !get {
		//该用户不存在
		return nil, err
	}
	//验证用户状态，
	//用户密码校对
	if !utils.BcryptCheck(req.Pass, userInfo.PassWord) {
		log.Log().Msg("密码错误")
		return nil, errors.New("密码错误")
	}
	//更新用户最近登录时间
	//
	return userInfo, nil
}

func (u *UserService) SaveUserToken(ctx *gin.Context, uid, token string) {
	_ = dao.SaveUserToken(ctx, uid, token)
}

func (u *UserService) DeleteUserToken(ctx context.Context, key string) {
	dao.DeleteUserToken(ctx, key)
}
