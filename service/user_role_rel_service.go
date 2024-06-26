package service

import (
	"context"
	entity "github.com/lawyer/commons/entity"
	"github.com/lawyer/repo"
)

// UserRoleRelRepo userRoleRel repository
type UserRoleRelRepo interface {
	SaveUserRoleRel(ctx context.Context, userID string, roleID int) (err error)
	GetUserRoleRelList(ctx context.Context, userIDs []string) (userRoleRelList []*entity.UserRoleRel, err error)
	GetUserRoleRelListByRoleID(ctx context.Context, roleIDs []int) (
		userRoleRelList []*entity.UserRoleRel, err error)
	GetUserRoleRel(ctx context.Context, userID string) (rolePowerRel *entity.UserRoleRel, exist bool, err error)
}

type UserRoleRelService struct {
}

func NewUserRoleRelService() *UserRoleRelService {
	return &UserRoleRelService{}
}

func (us *UserRoleRelService) SaveUserRole(ctx context.Context, userID string, roleID int) (err error) {
	return repo.UserRoleRelRepo.SaveUserRoleRel(ctx, userID, roleID)
}

// GetUserRoleMapping get user role mapping
func (us *UserRoleRelService) GetUserRoleMapping(ctx context.Context, userIDs []string) (
	userRoleMapping map[string]*entity.Role, err error) {
	userRoleMapping = make(map[string]*entity.Role, 0)
	roleMapping, err := RoleServicer.GetRoleMapping(ctx)
	if err != nil {
		return userRoleMapping, err
	}
	if len(roleMapping) == 0 {
		return userRoleMapping, nil
	}

	relMapping, err := us.GetUserRoleRelMapping(ctx, userIDs)
	if err != nil {
		return userRoleMapping, err
	}

	// default role is user
	defaultRole := roleMapping[1]
	for _, userID := range userIDs {
		roleID, ok := relMapping[userID]
		if !ok {
			userRoleMapping[userID] = defaultRole
			continue
		}
		userRoleMapping[userID] = roleMapping[roleID]
		if userRoleMapping[userID] == nil {
			userRoleMapping[userID] = defaultRole
		}
	}
	return userRoleMapping, nil
}

// GetUserRoleRelMapping get user role rel mapping
func (us *UserRoleRelService) GetUserRoleRelMapping(ctx context.Context, userIDs []string) (
	userRoleRelMapping map[string]int, err error) {
	userRoleRelMapping = make(map[string]int, 0)

	relList, err := repo.UserRoleRelRepo.GetUserRoleRelList(ctx, userIDs)
	if err != nil {
		return userRoleRelMapping, err
	}

	for _, rel := range relList {
		userRoleRelMapping[rel.UserID] = rel.RoleID
	}
	return userRoleRelMapping, nil
}

// GetUserRole get user role
func (us *UserRoleRelService) GetUserRole(ctx context.Context, userID string) (roleID int, err error) {
	rolePowerRel, exist, err := repo.UserRoleRelRepo.GetUserRoleRel(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !exist {
		// set default role
		return 1, nil
	}
	return rolePowerRel.RoleID, nil
}

// GetUserByRoleID get user by role id
func (us *UserRoleRelService) GetUserByRoleID(ctx context.Context, roleIDs []int) (rel []*entity.UserRoleRel, err error) {
	rolePowerRels, err := repo.UserRoleRelRepo.GetUserRoleRelListByRoleID(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	return rolePowerRels, nil
}
