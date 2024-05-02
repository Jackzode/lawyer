package controller_admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/middleware"
	services "github.com/lawyer/service"
)

// UserAdminController user controller
type UserAdminController struct {
}

// NewUserAdminController new controller
func NewUserAdminController() *UserAdminController {
	return &UserAdminController{}
}

// UpdateUserStatus update user
// @Summary update user
// @Description update user
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.UpdateUserStatusReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user/status [put]
func (uc *UserAdminController) UpdateUserStatus(ctx *gin.Context) {

	req := &schema.UpdateUserStatusReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	err := services.UserAdminServicer.UpdateUserStatus(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateUserRole update user role
// @Summary update user role
// @Description update user role
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.UpdateUserRoleReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user/role [put]
func (uc *UserAdminController) UpdateUserRole(ctx *gin.Context) {
	req := &schema.UpdateUserRoleReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	err := services.UserAdminServicer.UpdateUserRole(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// AddUser add user
// @Summary add user
// @Description add user
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.AddUserReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user [post]
func (uc *UserAdminController) AddUser(ctx *gin.Context) {
	req := &schema.AddUserReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	err := services.UserAdminServicer.AddUser(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// AddUsers add users
// @Summary add users
// @Description add users
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.AddUsersReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/users [post]
func (uc *UserAdminController) AddUsers(ctx *gin.Context) {
	req := &schema.AddUsersReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := services.UserAdminServicer.AddUsers(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateUserPassword update user password
// @Summary update user password
// @Description update user password
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.UpdateUserPasswordReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user/password [put]
func (uc *UserAdminController) UpdateUserPassword(ctx *gin.Context) {
	req := &schema.UpdateUserPasswordReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	err := services.UserAdminServicer.UpdateUserPassword(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetUserPage get user page
// @Summary get user page
// @Description get user page
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param page query int false "page size"
// @Param page_size query int false "page size"
// @Param query query string false "search query: email, username or id:[id]"
// @Param staff query bool false "staff user"
// @Param status query string false "user status" Enums(suspended, deleted, inactive)
// @Success 200 {object} handler.RespBody{data=pager.PageModel{records=[]schema.GetUserPageResp}}
// @Router /answer/admin/api/users/page [get]
func (uc *UserAdminController) GetUserPage(ctx *gin.Context) {
	req := &schema.GetUserPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := services.UserAdminServicer.GetUserPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetUserActivation get user activation
// @Summary get user activation
// @Description get user activation
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param user_id query string true "user id"
// @Success 200 {object} handler.RespBody{data=schema.GetUserActivationResp}
// @Router /answer/admin/api/user/activation [get]
func (uc *UserAdminController) GetUserActivation(ctx *gin.Context) {
	req := &schema.GetUserActivationReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := services.UserAdminServicer.GetUserActivation(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// SendUserActivation send user activation
// @Summary send user activation
// @Description send user activation
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SendUserActivationReq true "SendUserActivationReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/users/activation [post]
func (uc *UserAdminController) SendUserActivation(ctx *gin.Context) {
	req := &schema.SendUserActivationReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	err := services.UserAdminServicer.SendUserActivation(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
