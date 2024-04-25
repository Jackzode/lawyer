package auth

import (
	"context"
	"encoding/json"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// AuthRepo auth repository
type AuthRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewAuthRepo new repository
func NewAuthRepo() *AuthRepo {
	return &AuthRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// GetUserCacheInfo get user cache info
func (ar *AuthRepo) GetUserInfoFromCache(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error) {
	userInfoCache := ar.Cache.Get(ctx, constant.UserTokenCacheKey+accessToken).String()
	if userInfoCache == "" {
		return nil, nil
	}
	userInfo = &entity.UserCacheInfo{}
	_ = json.Unmarshal([]byte(userInfoCache), userInfo)
	return userInfo, nil
}

// SetUserCacheInfo set user cache info
func (ar *AuthRepo) SetUserCacheInfo(ctx context.Context,
	accessToken, visitToken string, userInfo *entity.UserCacheInfo) (err error) {
	userInfo.VisitToken = visitToken
	userInfoCache, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	err = ar.Cache.Set(ctx, constant.UserTokenCacheKey+accessToken,
		string(userInfoCache), constant.UserTokenCacheTime).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if err := ar.AddUserTokenMapping(ctx, userInfo.UserID, accessToken); err != nil {
		log.Error(err)
	}
	if len(visitToken) == 0 {
		return nil
	}
	if err = ar.Cache.Set(ctx, constant.UserVisitTokenCacheKey+visitToken,
		accessToken, constant.UserTokenCacheTime).Err(); err != nil {
		log.Error(err)
	}
	return nil
}

// GetUserVisitCacheInfo get user visit cache info
func (ar *AuthRepo) GetUserVisitCacheInfo(ctx context.Context, visitToken string) (accessToken string, err error) {
	accessToken = ar.Cache.Get(ctx, constant.UserVisitTokenCacheKey+visitToken).String()
	if accessToken == "" {
		return "", nil
	}
	return accessToken, nil
}

// RemoveUserCacheInfo remove user cache info
func (ar *AuthRepo) RemoveUserCacheInfo(ctx context.Context, accessToken string) (err error) {
	err = ar.Cache.Del(ctx, constant.UserTokenCacheKey+accessToken).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// RemoveUserVisitCacheInfo remove visit token cache
func (ar *AuthRepo) RemoveUserVisitCacheInfo(ctx context.Context, visitToken string) (err error) {
	err = ar.Cache.Del(ctx, constant.UserVisitTokenCacheKey+visitToken).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// SetUserStatus set user status
func (ar *AuthRepo) SetUserStatus(ctx context.Context, userID string, userInfo *entity.UserCacheInfo) (err error) {
	userInfoCache, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	err = ar.Cache.Set(ctx, constant.UserStatusCacheKey+userID,
		string(userInfoCache), constant.UserStatusChangedCacheTime).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// GetUserStatus get user status
func (ar *AuthRepo) GetUserStatusFromCache(ctx context.Context, userID string) (userInfo *entity.UserCacheInfo, err error) {
	userInfoCache := ar.Cache.Get(ctx, constant.UserStatusCacheKey+userID).String()
	if userInfoCache == "" {
		return nil, nil
	}
	userInfo = &entity.UserCacheInfo{}
	_ = json.Unmarshal([]byte(userInfoCache), userInfo)
	return userInfo, nil
}

// RemoveUserStatus remove user status
func (ar *AuthRepo) RemoveUserStatus(ctx context.Context, userID string) (err error) {
	err = ar.Cache.Del(ctx, constant.UserStatusCacheKey+userID).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// GetAdminUserCacheInfo get admin user cache info
func (ar *AuthRepo) GetAdminUserCacheInfo(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error) {
	userInfoCache := ar.Cache.Get(ctx, constant.AdminTokenCacheKey+accessToken).String()
	if userInfoCache == "" {
		return nil, nil
	}
	userInfo = &entity.UserCacheInfo{}
	_ = json.Unmarshal([]byte(userInfoCache), userInfo)
	return userInfo, nil
}

// SetAdminUserCacheInfo set admin user cache info
func (ar *AuthRepo) SetAdminUserCacheInfo(ctx context.Context, accessToken string, userInfo *entity.UserCacheInfo) (err error) {
	userInfoCache, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}

	err = ar.Cache.Set(ctx, constant.AdminTokenCacheKey+accessToken, string(userInfoCache),
		constant.AdminTokenCacheTime).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// RemoveAdminUserCacheInfo remove admin user cache info
func (ar *AuthRepo) RemoveAdminUserCacheInfo(ctx context.Context, accessToken string) (err error) {
	err = ar.Cache.Del(ctx, constant.AdminTokenCacheKey+accessToken).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// AddUserTokenMapping add user token mapping
func (ar *AuthRepo) AddUserTokenMapping(ctx context.Context, userID, accessToken string) (err error) {
	key := constant.UserTokenMappingCacheKey + userID
	resp := ar.Cache.Get(ctx, key).String()

	mapping := make(map[string]bool, 0)
	if len(resp) > 0 {
		_ = json.Unmarshal([]byte(resp), &mapping)
	}
	mapping[accessToken] = true
	content, _ := json.Marshal(mapping)
	return ar.Cache.Set(ctx, key, string(content), constant.UserTokenCacheTime).Err()
}

// RemoveUserTokens Log out all users under this user id
func (ar *AuthRepo) RemoveUserTokens(ctx context.Context, userID string, remainToken string) {
	key := constant.UserTokenMappingCacheKey + userID
	resp := ar.Cache.Get(ctx, key).String()
	if resp == "" {
		return
	}
	mapping := make(map[string]bool, 0)
	if len(resp) > 0 {
		_ = json.Unmarshal([]byte(resp), &mapping)
		log.Debugf("find %d user tokens by user id %s", len(mapping), userID)
	}

	for token := range mapping {
		if token == remainToken {
			continue
		}
		if err := ar.RemoveUserCacheInfo(ctx, token); err != nil {
			log.Error(err)
		} else {
			log.Debugf("del user %s token success")
		}
	}
	if err := ar.RemoveUserStatus(ctx, userID); err != nil {
		log.Error(err)
	}
	if err := ar.Cache.Del(ctx, key).Err(); err != nil {
		log.Error(err)
	}
}
