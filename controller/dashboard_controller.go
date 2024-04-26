package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/service"
)

type DashboardController struct {
	dashboardService service.DashboardService
}

// NewDashboardController new controller
func NewDashboardController(
	dashboardService service.DashboardService,
) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

func (ac *DashboardController) DashboardInfo(ctx *gin.Context) {
	info, err := ac.dashboardService.Statistical(ctx)
	handler.HandleResponse(ctx, err, gin.H{
		"info": info,
	})
}
