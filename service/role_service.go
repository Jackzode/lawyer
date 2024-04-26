package service

import (
	"context"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/repo"

	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/schema"
)

const (
	// Since there is currently no need to edit roles to add roles and other operations,
	// the current role information is translated directly.
	// Later on, when the relevant ability is available, it can be adjusted by the user himself.

	RoleUserID      = 1
	RoleAdminID     = 2
	RoleModeratorID = 3

	roleUserName      = "User"
	roleAdminName     = "Admin"
	roleModeratorName = "Moderator"

	trRoleNameUser      = "role.name.user"
	trRoleNameAdmin     = "role.name.admin"
	trRoleNameModerator = "role.name.moderator"

	trRoleDescriptionUser      = "role.description.user"
	trRoleDescriptionAdmin     = "role.description.admin"
	trRoleDescriptionModerator = "role.description.moderator"
)

// RoleRepo role repository
type RoleRepo interface {
	GetRoleAllList(ctx context.Context) (roles []*entity.Role, err error)
	GetRoleAllMapping(ctx context.Context) (roleMapping map[int]*entity.Role, err error)
}

// RoleServicer user service
type RoleService struct {
}

func NewRoleService() *RoleService {
	return &RoleService{}
}

// GetRoleList get role list all
func (rs *RoleService) GetRoleList(ctx context.Context) (resp []*schema.GetRoleResp, err error) {
	roles, err := repo.RoleRepo.GetRoleAllList(ctx)
	if err != nil {
		return
	}

	for _, role := range roles {
		rs.translateRole(ctx, role)
	}

	resp = []*schema.GetRoleResp{}
	_ = copier.Copy(&resp, roles)
	return
}

func (rs *RoleService) GetRoleMapping(ctx context.Context) (roleMapping map[int]*entity.Role, err error) {
	return repo.RoleRepo.GetRoleAllMapping(ctx)
}

func (rs *RoleService) translateRole(ctx context.Context, role *entity.Role) {
	switch role.Name {
	case roleUserName:
		role.Name = translator.Tr(utils.GetLangByCtx(ctx), trRoleNameUser)
		role.Description = translator.Tr(utils.GetLangByCtx(ctx), trRoleDescriptionUser)
	case roleAdminName:
		role.Name = translator.Tr(utils.GetLangByCtx(ctx), trRoleNameAdmin)
		role.Description = translator.Tr(utils.GetLangByCtx(ctx), trRoleDescriptionAdmin)
	case roleModeratorName:
		role.Name = translator.Tr(utils.GetLangByCtx(ctx), trRoleNameModerator)
		role.Description = translator.Tr(utils.GetLangByCtx(ctx), trRoleDescriptionModerator)
	}
}
