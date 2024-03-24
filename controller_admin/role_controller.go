package controller_admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	services "github.com/lawyer/initServer/initServices"
)

// RoleController role controller
type RoleController struct {
}

// NewRoleController new controller
func NewRoleController() *RoleController {
	return &RoleController{}
}

// GetRoleList get role list
// @Summary get role list
// @Description get role list
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=[]schema.GetRoleResp}
// @Router /answer/admin/api/roles [get]
func (rc *RoleController) GetRoleList(ctx *gin.Context) {
	req := &schema.GetRoleResp{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	resp, err := services.RoleService.GetRoleList(ctx)
	handler.HandleResponse(ctx, err, resp)
}
