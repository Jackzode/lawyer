package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/service/dashboard"
)

type DashboardController struct {
	dashboardService dashboard.DashboardService
}

// NewDashboardController new controller
func NewDashboardController(
	dashboardService dashboard.DashboardService,
) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

// DashboardInfo godoc
// @Summary DashboardInfo
// @Description DashboardInfo
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /answer/admin/api/dashboard [get]
// @Success 200 {object} handler.RespBody
func (ac *DashboardController) DashboardInfo(ctx *gin.Context) {
	info, err := ac.dashboardService.Statistical(ctx)
	handler.HandleResponse(ctx, err, gin.H{
		"info": info,
	})
}
