package role

import (
	"context"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/lawyer/initServer/initServices"
)

// RolePowerRelRepo rolePowerRel repository
type RolePowerRelRepo interface {
	GetRolePowerTypeList(ctx context.Context, roleID int) (powers []string, err error)
}

// RolePowerRelService user service
type RolePowerRelService struct {
}

// NewRolePowerRelService new role power rel service
func NewRolePowerRelService() *RolePowerRelService {
	return &RolePowerRelService{}
}

// GetRolePowerList get role power list
func (rs *RolePowerRelService) GetRolePowerList(ctx context.Context, roleID int) (powers []string, err error) {
	return repo.RolePowerRelRepo.GetRolePowerTypeList(ctx, roleID)
}

// GetUserPowerList get  list all
func (rs *RolePowerRelService) GetUserPowerList(ctx context.Context, userID string) (powers []string, err error) {
	roleID, err := services.UserRoleRelService.GetUserRole(ctx, userID)
	if err != nil {
		return nil, err
	}
	return repo.RolePowerRelRepo.GetRolePowerTypeList(ctx, roleID)
}
