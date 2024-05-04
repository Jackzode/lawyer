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
)

// AuthRepo auth repository
// 根据token或者uid操作缓存(key)，value是userCacheInfo
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
func (ar *AuthRepo) SetUserRegisterInfoByEmail(ctx context.Context, user *entity.User) (err error) {
	userInfo, err := json.Marshal(user)
	if err != nil {
		return err
	}
	err = ar.Cache.Set(ctx, constant.UserRegisterInfoKey+user.EMail,
		string(userInfo), constant.UserRegisterInfoTime).Err()
	return err
}

// 以uid为key， userCacheInfo为value，存到缓存里
func (ar *AuthRepo) SetUserCacheInfoByUid(ctx context.Context, userID string, userInfo *entity.UserCacheInfo) (err error) {
	userInfoCache, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	err = ar.Cache.Set(ctx, constant.UserCacheInfoKey+userID,
		string(userInfoCache), constant.UserCacheInfoChangedCacheTime).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// GetUserStatus get user status
// 根据uid从缓存中获取userCacheInfo
func (ar *AuthRepo) GetUserCacheInfoByUid(ctx context.Context, userID string) (userInfo *entity.UserCacheInfo, err error) {
	userInfoCache := ar.Cache.Get(ctx, constant.UserCacheInfoKey+userID).String()
	if userInfoCache == "" {
		return nil, nil
	}
	userInfo = &entity.UserCacheInfo{}
	_ = json.Unmarshal([]byte(userInfoCache), userInfo)
	return userInfo, nil
}

// RemoveUserStatus remove user status
// 根据uid删除userCacheInfo
func (ar *AuthRepo) RemoveUserCacheInfoByUid(ctx context.Context, userID string) (err error) {
	err = ar.Cache.Del(ctx, constant.UserCacheInfoKey+userID).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// GetUserCacheInfo get user cache info
// 根据acc token获取userCacheInfo
func (ar *AuthRepo) GetUserInfoFromCacheByToken(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error) {
	userInfoCache := ar.Cache.Get(ctx, constant.UserTokenCacheKey+accessToken).String()
	if userInfoCache == "" {
		return nil, nil
	}
	userInfo = &entity.UserCacheInfo{}
	_ = json.Unmarshal([]byte(userInfoCache), userInfo)
	return userInfo, nil
}

// SetUserCacheInfoByToken set user cache info
// accessToken指向user cache info， visit token指向acc token
// 中间有个AddUserTokenMapping没明白
func (ar *AuthRepo) SetUserCacheInfoByToken(ctx context.Context, accessToken string, userInfo *entity.UserCacheInfo) (err error) {

	userInfoCache, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	err = ar.Cache.Set(ctx, constant.UserTokenCacheKey+accessToken,
		string(userInfoCache), constant.UserTokenCacheTime).Err()
	return err
}

func (ar *AuthRepo) RemoveUserCacheInfoByToken(ctx context.Context, accessToken string) (err error) {
	err = ar.Cache.Del(ctx, constant.UserTokenCacheKey+accessToken).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// 检查visit token是否存在（visit token里存的是acc token）
//func (ar *AuthRepo) GetUserVisitCacheInfo(ctx context.Context, visitToken string) (accessToken string, err error) {
//	accessToken = ar.Cache.Get(ctx, constant.UserVisitTokenCacheKey+visitToken).String()
//	if accessToken == "" {
//		return "", nil
//	}
//	return accessToken, nil
//}

//func (ar *AuthRepo) RemoveUserVisitCacheInfo(ctx context.Context, visitToken string) (err error) {
//	err = ar.Cache.Del(ctx, constant.UserVisitTokenCacheKey+visitToken).Err()
//	if err != nil {
//		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
//	}
//	return nil
//}

//func (ar *AuthRepo) GetAdminUserCacheInfo(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error) {
//	userInfoCache := ar.Cache.Get(ctx, constant.AdminTokenCacheKey+accessToken).String()
//	if userInfoCache == "" {
//		return nil, nil
//	}
//	userInfo = &entity.UserCacheInfo{}
//	_ = json.Unmarshal([]byte(userInfoCache), userInfo)
//	return userInfo, nil
//}

//func (ar *AuthRepo) SetAdminUserCacheInfo(ctx context.Context, accessToken string, userInfo *entity.UserCacheInfo) (err error) {
//	userInfoCache, err := json.Marshal(userInfo)
//	if err != nil {
//		return err
//	}
//
//	err = ar.Cache.Set(ctx, constant.AdminTokenCacheKey+accessToken, string(userInfoCache),
//		constant.AdminTokenCacheTime).Err()
//	if err != nil {
//		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
//	}
//	return nil
//}

// RemoveAdminUserCacheInfo remove admin user cache info
//func (ar *AuthRepo) RemoveAdminUserCacheInfo(ctx context.Context, accessToken string) (err error) {
//	err = ar.Cache.Del(ctx, constant.AdminTokenCacheKey+accessToken).Err()
//	if err != nil {
//		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
//	}
//	return nil
//}

//func (ar *AuthRepo) AddUserTokenMapping(ctx context.Context, userID, accessToken string) (err error) {
//	key := constant.UserTokenMappingCacheKey + userID
//	resp := ar.Cache.Get(ctx, key).String()
//
//	mapping := make(map[string]bool, 0)
//	if len(resp) > 0 {
//		_ = json.Unmarshal([]byte(resp), &mapping)
//	}
//	mapping[accessToken] = true
//	content, _ := json.Marshal(mapping)
//	return ar.Cache.Set(ctx, key, string(content), constant.UserTokenCacheTime).Err()
//}
//
//func (ar *AuthRepo) RemoveUserTokens(ctx context.Context, userID string, remainToken string) {
//	key := constant.UserTokenMappingCacheKey + userID
//	resp := ar.Cache.Get(ctx, key).String()
//	if resp == "" {
//		return
//	}
//	mapping := make(map[string]bool, 0)
//	if len(resp) > 0 {
//		_ = json.Unmarshal([]byte(resp), &mapping)
//		log.Debugf("find %d user tokens by user id %s", len(mapping), userID)
//	}
//
//	for token := range mapping {
//		if token == remainToken {
//			continue
//		}
//		if err := ar.RemoveUserCacheInfoByToken(ctx, token); err != nil {
//			log.Error(err)
//		} else {
//			log.Debugf("del user %s token success")
//		}
//	}
//	if err := ar.RemoveUserCacheInfoByUid(ctx, userID); err != nil {
//		log.Error(err)
//	}
//	if err := ar.Cache.Del(ctx, key).Err(); err != nil {
//		log.Error(err)
//	}
//}
