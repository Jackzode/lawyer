package service

import (
	"context"
	"github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/repo"
	"time"
)

// ValidationStrategy
// true pass
// false need captcha
func (cs *CaptchaService) ValidationStrategy(ctx context.Context, IP, actionType string) bool {
	//在缓存中查询最近6min的操作
	info, err := repo.CaptchaRepo.GetActionType(ctx, IP, actionType)
	if err != nil {
		glog.Logger.Error(err.Error())
		return false
	}
	// If no operation previously, it is considered to be the first operation
	//第一次操作不需要验证码之类的
	if info == nil {
		return true
	}
	switch actionType {
	case entity.CaptchaActionEmail:
		return cs.CaptchaActionEmail(ctx, IP, info)
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
	return false
}

func (cs *CaptchaService) CaptchaActionPassword(ctx context.Context, unit string, actioninfo *entity.ActionRecordInfo) bool {
	setNum := 3
	setTime := int64(60 * 30) //seconds
	now := time.Now().Unix()
	if now-actioninfo.LastTime <= setTime && actioninfo.Num >= setNum {
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
