package service

import (
	"context"
	"fmt"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/repo"
	"image/color"
	"strings"
	"time"

	"github.com/lawyer/commons/schema"
	"github.com/mojocn/base64Captcha"
	"github.com/segmentfault/pacman/errors"
)

// CaptchaRepo captcha repository
type CaptchaRepo interface {
	SetCaptcha(ctx context.Context, key, captcha string) (err error)
	GetCaptcha(ctx context.Context, key string) (captcha string, err error)
	DelCaptcha(ctx context.Context, key string) (err error)
	SetActionType(ctx context.Context, unit, actionType, config string, amount int) (err error)
	GetActionType(ctx context.Context, unit, actionType string) (actioninfo *entity.ActionRecordInfo, err error)
	DelActionType(ctx context.Context, unit, actionType string) (err error)
}

// CaptchaServicer kit service
type CaptchaService struct {
}

// NewCaptchaService captcha service
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{}
}

// ActionRecord action record
// 针对不同的action，选择用户唯一标识
func (cs *CaptchaService) GetActionRecordUnit(ctx context.Context, req *schema.ActionRecordReq) string {

	unit := req.IP
	switch req.Action {
	case entity.CaptchaActionEditUserinfo:
		unit = req.UserID
	case entity.CaptchaActionQuestion:
		unit = req.UserID
	case entity.CaptchaActionAnswer:
		unit = req.UserID
	case entity.CaptchaActionComment:
		unit = req.UserID
	case entity.CaptchaActionEdit:
		unit = req.UserID
	case entity.CaptchaActionInvitationAnswer:
		unit = req.UserID
	case entity.CaptchaActionSearch:
		if req.UserID != "" {
			unit = req.UserID
		}
	case entity.CaptchaActionReport:
		unit = req.UserID
	case entity.CaptchaActionDelete:
		unit = req.UserID
	case entity.CaptchaActionVote:
		unit = req.UserID
	}
	return unit
}

// 生成验证码
func (cs *CaptchaService) UserRegisterCaptcha(ctx context.Context) (resp *schema.ActionRecordResp, err error) {
	resp = &schema.ActionRecordResp{}
	resp.CaptchaID, resp.CaptchaImg, err = cs.GenerateCaptcha(ctx)
	resp.Verify = true
	return
}

// 校对验证码
func (cs *CaptchaService) UserRegisterVerifyCaptcha(ctx context.Context, id, VerifyValue string) bool {
	if id == "" || VerifyValue == "" {
		return false
	}
	pass, err := cs.VerifyCaptcha(ctx, id, VerifyValue)
	if err != nil {
		return false
	}
	return pass
}

// ActionRecordVerifyCaptcha
// Verify that you need to enter a CAPTCHA, and that the CAPTCHA is correct
func (cs *CaptchaService) ActionRecordVerifyCaptcha(ctx context.Context, actionType, IP, CaptchaId, CaptchaCode string) bool {
	NeedVerification := cs.ValidationStrategy(ctx, IP, actionType)
	fmt.Println("NeedVerification； ", NeedVerification)
	//需要验证码
	if NeedVerification {
		return cs.UserRegisterVerifyCaptcha(ctx, CaptchaId, CaptchaCode)
	}
	//直接通过
	return true
}

// 对每个用户的每个操作记录一下，为的是防止用户过度操作
func (cs *CaptchaService) ActionRecordAdd(ctx context.Context, actionType string, unit string) (int, error) {
	info, err := repo.CaptchaRepo.GetActionType(ctx, unit, actionType)
	if err != nil {
		glog.Slog.Error(err)
		return 0, err
	}
	amount := 1
	if info != nil {
		amount = info.Num + 1
	}
	err = repo.CaptchaRepo.SetActionType(ctx, unit, actionType, "", amount)
	if err != nil {
		return 0, err
	}
	return amount, nil
}

func (cs *CaptchaService) ActionRecordDel(ctx context.Context, actionType string, unit string) {
	err := repo.CaptchaRepo.DelActionType(ctx, unit, actionType)
	if err != nil {
		glog.Slog.Error(err)
	}
}

// GenerateCaptcha generate captcha
func (cs *CaptchaService) GenerateCaptcha(ctx context.Context) (key, captchaBase64 string, err error) {
	driverString := base64Captcha.DriverString{
		Height:          60,
		Width:           200,
		NoiseCount:      0,
		ShowLineOptions: 2 | 4,
		Length:          4,
		Source:          "1234567890qwertyuioplkjhgfdsazxcvbnm",
		BgColor:         &color.RGBA{R: 211, G: 211, B: 211, A: 0},
		Fonts:           []string{"wqy-microhei.ttc"},
	}
	driver := driverString.ConvertFonts()

	id, content, answer := driver.GenerateIdQuestionAnswer()
	item, err := driver.DrawCaptcha(content)
	if err != nil {
		return "", "", errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	err = repo.CaptchaRepo.SetCaptcha(ctx, id, answer)
	if err != nil {
		return "", "", err
	}

	captchaBase64 = item.EncodeB64string()
	return id, captchaBase64, nil
}

// VerifyCaptcha generate captcha 拿到验证码答案，在删除验证码key，再和端上入参进行对比
func (cs *CaptchaService) VerifyCaptcha(ctx context.Context, key, captcha string) (isCorrect bool, err error) {
	realCaptcha, err := repo.CaptchaRepo.GetCaptcha(ctx, key)
	if err != nil {
		glog.Slog.Error("VerifyCaptcha GetCaptcha Error", err.Error())
		return false, nil
	}
	err = repo.CaptchaRepo.DelCaptcha(ctx, key)
	if err != nil {
		glog.Slog.Error("VerifyCaptcha DelCaptcha Error", err.Error())
		return false, nil
	}
	return strings.TrimSpace(captcha) == realCaptcha, nil
}

// 判断一下当前操作是否需要验证码，true需要验证，false直接跳过
func (cs *CaptchaService) ValidationStrategy(ctx context.Context, IP, actionType string) bool {
	//在缓存中查询最近操作
	info, err := repo.CaptchaRepo.GetActionType(ctx, IP, actionType)
	fmt.Println("info= ", info, err)
	if err != nil {
		glog.Klog.Error(err.Error())
		return true
	}
	// If no operation previously, it is considered to be the first operation
	//最近没有操作记录，不需要验证码
	if info == nil {
		return false
	}
	switch actionType {
	//邮件类型每次都需要验证，直接返回true
	case entity.CaptchaActionEmail:
		return cs.CaptchaActionEmail(ctx, IP, info)
		//下面的几种类型还没看
	case entity.CaptchaActionPassword:
		return cs.CaptchaActionPassword(ctx, IP, info)
	case entity.CaptchaActionEditUserinfo:
		return cs.CaptchaActionEditUserinfo(ctx, IP, info)
	case entity.CaptchaActionQuestion:
		return cs.CaptchaActionQuestion(ctx, IP, info)
	case entity.CaptchaActionAnswer:
		return cs.CaptchaActionAnswer(ctx, IP, info)
	case entity.CaptchaActionComment:
		return cs.CaptchaActionComment(ctx, IP, info)
	case entity.CaptchaActionEdit:
		return cs.CaptchaActionEdit(ctx, IP, info)
	case entity.CaptchaActionInvitationAnswer:
		return cs.CaptchaActionInvitationAnswer(ctx, IP, info)
	case entity.CaptchaActionSearch:
		return cs.CaptchaActionSearch(ctx, IP, info)
	case entity.CaptchaActionReport:
		return cs.CaptchaActionReport(ctx, IP, info)
	case entity.CaptchaActionDelete:
		return cs.CaptchaActionDelete(ctx, IP, info)
	case entity.CaptchaActionVote:
		return cs.CaptchaActionVote(ctx, IP, info)

	}
	//actionType not found
	return false
}

func (cs *CaptchaService) CaptchaActionEmail(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	// You need a verification code every time
	return true
}

func (cs *CaptchaService) CaptchaActionPassword(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 3
	setTime := int64(60 * 30) //seconds
	now := time.Now().Unix()
	if now-actioninfo.LastTime <= setTime && actioninfo.Num <= setNum {
		return false
	}
	if now-actioninfo.LastTime != 0 && now-actioninfo.LastTime > setTime {
		repo.CaptchaRepo.SetActionType(ctx, unit, entity.CaptchaActionPassword, "", 0)
	}
	return true
}

func (cs *CaptchaService) CaptchaActionEditUserinfo(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 3
	setTime := int64(60 * 30) //seconds
	now := time.Now().Unix()
	if now-actioninfo.LastTime <= setTime && actioninfo.Num >= setNum {
		return false
	}
	if now-actioninfo.LastTime != 0 && now-actioninfo.LastTime > setTime {
		repo.CaptchaRepo.SetActionType(ctx, unit, entity.CaptchaActionEditUserinfo, "", 0)
	}
	return true
}

func (cs *CaptchaService) CaptchaActionQuestion(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 10
	setTime := int64(5) //seconds
	now := time.Now().Unix()
	if now-actioninfo.LastTime <= setTime || actioninfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionAnswer(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 10
	setTime := int64(5) //seconds
	now := time.Now().Unix()
	if now-actioninfo.LastTime <= setTime || actioninfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionComment(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 30
	setTime := int64(1) //seconds
	now := time.Now().Unix()
	if now-actioninfo.LastTime <= setTime || actioninfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionEdit(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 10
	if actioninfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionInvitationAnswer(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 30
	if actioninfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionSearch(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	now := time.Now().Unix()
	setNum := 20
	setTime := int64(60) //seconds
	if now-int64(actioninfo.LastTime) <= setTime && actioninfo.Num >= setNum {
		return false
	}
	if now-actioninfo.LastTime > setTime {
		repo.CaptchaRepo.SetActionType(ctx, unit, entity.CaptchaActionSearch, "", 0)
	}
	return true
}

func (cs *CaptchaService) CaptchaActionReport(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 30
	setTime := int64(1) //seconds
	now := time.Now().Unix()
	if now-actioninfo.LastTime <= setTime || actioninfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionDelete(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 5
	setTime := int64(5) //seconds
	now := time.Now().Unix()
	if now-actioninfo.LastTime <= setTime || actioninfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionVote(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 40
	if actioninfo.Num >= setNum {
		return false
	}
	return true
}
