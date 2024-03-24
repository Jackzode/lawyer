package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	services "github.com/lawyer/initServer/initServices"
)

type DashboardController struct {
	//dashboardService dashboard.DashboardService
}

// NewDashboardController new controller
func NewDashboardController() *DashboardController {
	return &DashboardController{}
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
	info, err := services.DashboardService.Statistical(ctx)
	handler.HandleResponse(ctx, err, gin.H{
		"info": info,
	})
}
