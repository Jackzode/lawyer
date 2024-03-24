package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	services "github.com/lawyer/initServer/initServices"
	"github.com/lawyer/middleware"
	"github.com/lawyer/service/permission"
)

// NotificationController notification controller
type NotificationController struct {
	//notificationService *notification.NotificationService
	//rankService         *rank.RankService
}

// NewNotificationController new controller
func NewNotificationController() *NotificationController {
	return &NotificationController{}
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
	canList, err := services.RankService.CheckOperationPermissions(ctx, req.UserID, []string{
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

	resp, err := services.NotificationService.GetRedDot(ctx, req)
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
	canList, err := services.RankService.CheckOperationPermissions(ctx, req.UserID, []string{
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

	RedDot, err := services.NotificationService.ClearRedDot(ctx, req)
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
	err := services.NotificationService.ClearUnRead(ctx, userID, req.TypeStr)
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
	err := services.NotificationService.ClearIDUnRead(ctx, userID, req.ID)
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
	resp, err := services.NotificationService.GetNotificationPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
