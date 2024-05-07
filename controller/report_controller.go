package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/pkg/uid"
	"github.com/lawyer/service"
	"github.com/lawyer/service/permission"
	"github.com/segmentfault/pacman/errors"
)

// ReportController report controller
type ReportController struct {
	//reportService *service.ReportService
	//rankService   *service.RankService
	//actionService *service.CaptchaService
}

// NewReportController new controller
func NewReportController(
// reportService *service.ReportService,
// rankService *service.RankService,
// actionService *service.CaptchaService,
) *ReportController {
	return &ReportController{
		//reportService: reportService,
		//rankService:   rankService,
		//actionService: actionService,
	}
}

// AddReport add report
// @Summary add report
// @Description add report <br> source (question, answer, comment, user)
// @Security ApiKeyAuth
// @Tags Report
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.AddReportReq true "report"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/report [post]
func (rc *ReportController) AddReport(ctx *gin.Context) {
	req := &schema.AddReportReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)
	//req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.UserID = utils.GetUidFromTokenByCtx(ctx)
	//isAdmin := middleware.GetUserIsAdminModerator(ctx)
	//if !isAdmin {
	captchaPass := service.CaptchaServicer.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionReport, req.UserID, req.CaptchaID, req.CaptchaCode)
	if !captchaPass {
		handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), nil)
		return
	}

	can, err := service.RankServicer.CheckOperationPermission(ctx, req.UserID, permission.ReportAdd, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = service.ReportServicer.AddReport(ctx, req)
	service.CaptchaServicer.ActionRecordAdd(ctx, entity.CaptchaActionReport, req.UserID)
	handler.HandleResponse(ctx, err, nil)
}
