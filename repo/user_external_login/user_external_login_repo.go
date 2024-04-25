package user_external_login

import (
	"context"
	"encoding/json"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/lawyer/commons/schema"
	"github.com/segmentfault/pacman/errors"
)

type UserExternalLoginRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewUserExternalLoginRepo new repository
func NewUserExternalLoginRepo() *UserExternalLoginRepo {
	return &UserExternalLoginRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddUserExternalLogin add external login information
func (ur *UserExternalLoginRepo) AddUserExternalLogin(ctx context.Context, user *entity.UserExternalLogin) (err error) {
	_, err = ur.DB.Context(ctx).Insert(user)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateInfo update user info
func (ur *UserExternalLoginRepo) UpdateInfo(ctx context.Context, userInfo *entity.UserExternalLogin) (err error) {
	_, err = ur.DB.Context(ctx).ID(userInfo.ID).Update(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetByExternalID get by external ID
func (ur *UserExternalLoginRepo) GetByExternalID(ctx context.Context, provider, externalID string) (
	userInfo *entity.UserExternalLogin, exist bool, err error) {
	userInfo = &entity.UserExternalLogin{}
	exist, err = ur.DB.Context(ctx).Where("external_id = ?", externalID).Where("provider = ?", provider).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUserExternalLoginList get by external ID
func (ur *UserExternalLoginRepo) GetUserExternalLoginList(ctx context.Context, userID string) (
	resp []*entity.UserExternalLogin, err error) {
	resp = make([]*entity.UserExternalLogin, 0)
	err = ur.DB.Context(ctx).Where("user_id = ?", userID).Find(&resp)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// DeleteUserExternalLogin delete external user login info
func (ur *UserExternalLoginRepo) DeleteUserExternalLogin(ctx context.Context, userID, externalID string) (err error) {
	cond := &entity.UserExternalLogin{}
	_, err = ur.DB.Context(ctx).Where("user_id = ? AND external_id = ?", userID, externalID).Delete(cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SetCacheUserExternalLoginInfo cache user info for external login
func (ur *UserExternalLoginRepo) SetCacheUserExternalLoginInfo(
	ctx context.Context, key string, info *schema.ExternalLoginUserInfoCache) (err error) {
	cacheData, _ := json.Marshal(info)
	return ur.Cache.Set(ctx, constant.ConnectorUserExternalInfoCacheKey+key,
		string(cacheData), constant.ConnectorUserExternalInfoCacheTime).Err()
}

// GetCacheUserExternalLoginInfo cache user info for external login
func (ur *UserExternalLoginRepo) GetCacheUserExternalLoginInfo(
	ctx context.Context, key string) (info *schema.ExternalLoginUserInfoCache, err error) {
	res := ur.Cache.Get(ctx, constant.ConnectorUserExternalInfoCacheKey+key).String()
	if res == "" {
		return nil, nil
	}
	info = &schema.ExternalLoginUserInfoCache{}
	_ = json.Unmarshal([]byte(res), &info)
	return info, nil
}
