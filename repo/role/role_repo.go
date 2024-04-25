package role

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
)

// RoleRepo role repository
type RoleRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewRoleRepo new repository
func NewRoleRepo() *RoleRepo {
	return &RoleRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// GetRoleAllList get role list all
func (rr *RoleRepo) GetRoleAllList(ctx context.Context) (roleList []*entity.Role, err error) {
	roleList = make([]*entity.Role, 0)
	err = rr.DB.Context(ctx).Find(&roleList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetRoleAllMapping get role all mapping
func (rr *RoleRepo) GetRoleAllMapping(ctx context.Context) (roleMapping map[int]*entity.Role, err error) {
	roleList, err := rr.GetRoleAllList(ctx)
	if err != nil {
		return nil, err
	}
	roleMapping = make(map[int]*entity.Role, 0)
	for _, role := range roleList {
		roleMapping[role.ID] = role
	}
	return roleMapping, nil
}
