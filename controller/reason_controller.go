package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	services "github.com/lawyer/initServer/initServices"
)

// ReasonController answer controller
type ReasonController struct {
	//reasonService *reason.ReasonService
}

// NewReasonController new controller
func NewReasonController() *ReasonController {
	return &ReasonController{}
}

// Reasons godoc
// @Summary get reasons by object type and action
// @Description get reasons by object type and action
// @Tags reason
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param object_type query string true "object_type" Enums(question, answer, comment, user)
// @Param action query string true "action" Enums(status, close, flag, review)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/reasons [get]
// @Router /answer/admin/api/reasons [get]
func (rc *ReasonController) Reasons(ctx *gin.Context) {
	req := &schema.ReasonReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	reasons, err := services.ReasonService.GetReasons(ctx, *req)
	handler.HandleResponse(ctx, err, reasons)
}
