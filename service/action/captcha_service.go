package action

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	repo "github.com/lawyer/initServer/initRepo"
	"image/color"
	"strings"

	"github.com/lawyer/commons/schema"
	"github.com/mojocn/base64Captcha"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
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

// CaptchaService kit service
type CaptchaService struct {
}

// NewCaptchaService captcha service
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{}
}

// ActionRecord action record
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

func (cs *CaptchaService) UserRegisterCaptcha(ctx context.Context) (resp *schema.ActionRecordResp, err error) {
	resp = &schema.ActionRecordResp{}
	resp.CaptchaID, resp.CaptchaImg, err = cs.GenerateCaptcha(ctx)
	resp.Verify = true
	return
}

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
	verificationResult := cs.ValidationStrategy(ctx, IP, actionType)
	if !verificationResult {
		if CaptchaId == "" || CaptchaCode == "" {
			return false
		}
		pass, err := cs.VerifyCaptcha(ctx, CaptchaId, CaptchaCode)
		if err != nil {
			return false
		}
		return pass
	}
	return true
}

func (cs *CaptchaService) ActionRecordAdd(ctx context.Context, actionType string, unit string) (int, error) {
	info, err := repo.CaptchaRepo.GetActionType(ctx, unit, actionType)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
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

// VerifyCaptcha generate captcha
func (cs *CaptchaService) VerifyCaptcha(ctx context.Context, key, captcha string) (isCorrect bool, err error) {
	realCaptcha, err := repo.CaptchaRepo.GetCaptcha(ctx, key)
	if err != nil {
		log.Error("VerifyCaptcha GetCaptcha Error", err.Error())
		return false, nil
	}
	err = repo.CaptchaRepo.DelCaptcha(ctx, key)
	if err != nil {
		log.Error("VerifyCaptcha DelCaptcha Error", err.Error())
		return false, nil
	}
	return strings.TrimSpace(captcha) == realCaptcha, nil
}