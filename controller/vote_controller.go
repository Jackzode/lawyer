package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/middleware"
	"github.com/lawyer/pkg/uid"
	"github.com/lawyer/service"
	"github.com/segmentfault/pacman/errors"
)

// VoteController activity controller
type VoteController struct {
}

// NewVoteController new controller
func NewVoteController() *VoteController {
	return &VoteController{}
}

// VoteUp godoc
// @Summary vote up
// @Description add vote
// @Tags Activity
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.VoteReq true "vote"
// @Success 200 {object} handler.RespBody{data=schema.VoteResp}
// @Router /answer/api/v1/vote/up [post]
func (vc *VoteController) VoteUp(ctx *gin.Context) {
	req := &schema.VoteReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	can, needRank, err := service.RankServicer.CheckVotePermission(ctx, req.UserID, req.ObjectID, true)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		lang := utils.GetLang(ctx)
		msg := translator.TrWithData(lang, reason.NoEnoughRankToOperate, &schema.PermissionTrTplData{Rank: needRank})
		handler.HandleResponse(ctx, errors.Forbidden(reason.NoEnoughRankToOperate).WithMsg(msg), nil)
		return
	}

	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionVote, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	if !isAdmin {
		service.CaptchaServicer.ActionRecordAdd(ctx, entity.CaptchaActionVote, req.UserID)
	}
	resp, err := service.VoteServicer.VoteUp(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, schema.ErrTypeToast)
	} else {
		handler.HandleResponse(ctx, err, resp)
	}
}

// VoteDown godoc
// @Summary vote down
// @Description add vote
// @Tags Activity
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.VoteReq true "vote"
// @Success 200 {object} handler.RespBody{data=schema.VoteResp}
// @Router /answer/api/v1/vote/down [post]
func (vc *VoteController) VoteDown(ctx *gin.Context) {
	req := &schema.VoteReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)

	can, needRank, err := service.RankServicer.CheckVotePermission(ctx, req.UserID, req.ObjectID, false)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		lang := utils.GetLang(ctx)
		msg := translator.TrWithData(lang, reason.NoEnoughRankToOperate, &schema.PermissionTrTplData{Rank: needRank})
		handler.HandleResponse(ctx, errors.Forbidden(reason.NoEnoughRankToOperate).WithMsg(msg), nil)
		return
	}

	if !isAdmin {
		captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionVote, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(utils.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}
	if !isAdmin {
		service.CaptchaServicer.ActionRecordAdd(ctx, entity.CaptchaActionVote, req.UserID)
	}
	resp, err := service.VoteServicer.VoteDown(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, schema.ErrTypeToast)
	} else {
		handler.HandleResponse(ctx, err, resp)
	}
}

// UserVotes user votes
// @Summary get user personal votes
// @Description get user personal votes
// @Tags Activity
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page size"
// @Param page_size query int false "page size"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetVoteWithPageResp}}
// @Router /answer/api/v1/personal/vote/page [get]
func (vc *VoteController) UserVotes(ctx *gin.Context) {
	req := schema.GetVoteWithPageReq{}
	if handler.BindAndCheck(ctx, &req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := service.VoteServicer.ListUserVotes(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
