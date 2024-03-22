package role

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"

	"github.com/lawyer/service/role"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// userRoleRelRepo userRoleRel repository
type userRoleRelRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewUserRoleRelRepo new repository
func NewUserRoleRelRepo() role.UserRoleRelRepo {
	return &userRoleRelRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// SaveUserRoleRel save user role rel
func (ur *userRoleRelRepo) SaveUserRoleRel(ctx context.Context, userID string, roleID int) (err error) {
	_, err = ur.DB.Transaction(func(session *xorm.Session) (interface{}, error) {
		session = session.Context(ctx)
		item := &entity.UserRoleRel{UserID: userID}
		exist, err := session.Get(item)
		if err != nil {
			return nil, err
		}
		if exist {
			item.RoleID = roleID
			_, err = session.ID(item.ID).Update(item)
		} else {
			_, err = session.Insert(&entity.UserRoleRel{UserID: userID, RoleID: roleID})
		}
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUserRoleRelList get user role all
func (ur *userRoleRelRepo) GetUserRoleRelList(ctx context.Context, userIDs []string) (
	userRoleRelList []*entity.UserRoleRel, err error) {
	userRoleRelList = make([]*entity.UserRoleRel, 0)
	err = ur.DB.Context(ctx).In("user_id", userIDs).Find(&userRoleRelList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUserRoleRelListByRoleID get user role all by role id
func (ur *userRoleRelRepo) GetUserRoleRelListByRoleID(ctx context.Context, roleIDs []int) (
	userRoleRelList []*entity.UserRoleRel, err error) {
	userRoleRelList = make([]*entity.UserRoleRel, 0)
	err = ur.DB.Context(ctx).In("role_id", roleIDs).Find(&userRoleRelList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUserRoleRel get user role
func (ur *userRoleRelRepo) GetUserRoleRel(ctx context.Context, userID string) (
	rolePowerRel *entity.UserRoleRel, exist bool, err error) {
	rolePowerRel = &entity.UserRoleRel{}
	exist, err = ur.DB.Context(ctx).Where(builder.Eq{"user_id": userID}).Get(rolePowerRel)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
