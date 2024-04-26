package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/middleware"
	"github.com/lawyer/service"
	"github.com/lawyer/service/permission"
)

// NotificationController notification controller
type NotificationController struct {
	notificationService *service.NotificationService
	rankService         *service.RankService
}

// NewNotificationController new controller
func NewNotificationController(
	notificationService *service.NotificationService,
	rankService *service.RankService,
) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		rankService:         rankService,
	}
}

// GetRedDot
// @Summary GetRedDot
// @Description GetRedDot
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/status [get]
func (nc *NotificationController) GetRedDot(ctx *gin.Context) {
	req := &schema.GetRedDot{}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := nc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.QuestionAudit,
		permission.AnswerAudit,
		permission.TagAudit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanReviewQuestion = canList[0]
	req.CanReviewAnswer = canList[1]
	req.CanReviewTag = canList[2]

	resp, err := nc.notificationService.GetRedDot(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// ClearRedDot
// @Summary DelRedDot
// @Description DelRedDot
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.NotificationClearRequest true "NotificationClearRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/status [put]
func (nc *NotificationController) ClearRedDot(ctx *gin.Context) {
	req := &schema.NotificationClearRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := nc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.QuestionAudit,
		permission.AnswerAudit,
		permission.TagAudit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanReviewQuestion = canList[0]
	req.CanReviewAnswer = canList[1]
	req.CanReviewTag = canList[2]

	RedDot, err := nc.notificationService.ClearRedDot(ctx, req)
	handler.HandleResponse(ctx, err, RedDot)
}

// ClearUnRead
// @Summary ClearUnRead
// @Description ClearUnRead
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.NotificationClearRequest true "NotificationClearRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/read/state/all [put]
func (nc *NotificationController) ClearUnRead(ctx *gin.Context) {
	req := &schema.NotificationClearRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	userID := middleware.GetLoginUserIDFromContext(ctx)
	err := nc.notificationService.ClearUnRead(ctx, userID, req.TypeStr)
	handler.HandleResponse(ctx, err, gin.H{})
}

// ClearIDUnRead
// @Summary ClearUnRead
// @Description ClearUnRead
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.NotificationClearIDRequest true "NotificationClearIDRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/read/state [put]
func (nc *NotificationController) ClearIDUnRead(ctx *gin.Context) {
	req := &schema.NotificationClearIDRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	userID := middleware.GetLoginUserIDFromContext(ctx)
	err := nc.notificationService.ClearIDUnRead(ctx, userID, req.ID)
	handler.HandleResponse(ctx, err, gin.H{})
}

// GetList get notification list
// @Summary get notification list
// @Description get notification list
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page size"
// @Param page_size query int false "page size"
// @Param type query string true "type" Enums(inbox,achievement)
// @Param inbox_type query string true "inbox_type" Enums(all,posts,invites,votes)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/page [get]
func (nc *NotificationController) GetList(ctx *gin.Context) {
	req := &schema.NotificationSearch{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	resp, err := nc.notificationService.GetNotificationPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
