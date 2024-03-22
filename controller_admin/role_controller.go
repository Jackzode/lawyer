package controller_admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	service "github.com/lawyer/service/role"
)

// RoleController role controller
type RoleController struct {
	roleService *service.RoleService
}

// NewRoleController new controller
func NewRoleController(roleService *service.RoleService) *RoleController {
	return &RoleController{roleService: roleService}
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
	resp, err := rc.roleService.GetRoleList(ctx)
	handler.HandleResponse(ctx, err, resp)
}
