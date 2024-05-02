package service

import (
	"context"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/repo"
)

type IAuthRepo interface {
	SetUserCacheInfoByUid(ctx context.Context, userID string, userInfo *entity.UserCacheInfo) (err error)
	GetUserCacheInfoByUid(ctx context.Context, userID string) (userInfo *entity.UserCacheInfo, err error)
	RemoveUserCacheInfoByUid(ctx context.Context, userID string) (err error)
	GetUserInfoFromCacheByToken(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error)
	SetUserCacheInfoByToken(ctx context.Context, accessToken string, userInfo *entity.UserCacheInfo) (err error)
	RemoveUserCacheInfoByToken(ctx context.Context, accessToken string) (err error)
}

// AuthServicer kit service
// 操作缓存，根据uid或者token增删改userCacheInfo
type AuthService struct {
	r IAuthRepo
}

// NewAuthService email service
func NewAuthService() *AuthService {
	return &AuthService{
		repo.AuthRepo,
	}
}

func (as *AuthService) GetUserCacheInfo(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error) {
	userCacheInfo, err := as.r.GetUserInfoFromCacheByToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if userCacheInfo == nil {
		return nil, nil
	}
	//user表里包含了roleId,不需要再次请求了
	//userStatusInfo, _ := as.r.GetUserCacheInfoByUid(ctx, userCacheInfo.UserID)
	//if userStatusInfo != nil {
	//	userCacheInfo.UserStatus = userStatusInfo.UserStatus
	//	userCacheInfo.EmailStatus = userStatusInfo.EmailStatus
	//	userCacheInfo.RoleID = userStatusInfo.RoleID
	//	// update current user cache info
	//	err = as.r.SetUserCacheInfoByToken(ctx, accessToken, userCacheInfo)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	//todo  try to get user status from user center
	//uc, ok := plugin.GetUserCenter()
	//if ok && len(userCacheInfo.ExternalID) > 0 {
	//	if userStatus := uc.UserStatus(userCacheInfo.ExternalID); userStatus != plugin.UserStatusAvailable {
	//		userCacheInfo.UserStatus = int(userStatus)
	//	}
	//}
	return userCacheInfo, nil
}

// visit token 指向access token， acc token指向userCacheInfo
func (as *AuthService) SetUserCacheInfo(ctx context.Context, userInfo *entity.UserCacheInfo) (
	accessToken string, err error) {
	//accessToken = token.GenerateToken()
	accessToken, err = utils.CreateToken(userInfo.UserName, userInfo.UserID, userInfo.RoleID)
	if err != nil {
		return "", err
	}
	err = as.r.SetUserCacheInfoByToken(ctx, accessToken, userInfo)
	if err != nil {
		return "", err
	}
	return accessToken, err
}

func (as *AuthService) SetUserCacheInfoByUid(ctx context.Context, userInfo *entity.UserCacheInfo) (err error) {
	return as.r.SetUserCacheInfoByUid(ctx, userInfo.UserID, userInfo)
}

func (as *AuthService) RemoveUserCacheInfo(ctx context.Context, accessToken string) (err error) {
	return as.r.RemoveUserCacheInfoByToken(ctx, accessToken)
}

// 检查visit token是否存在（visit token里存的是acc token）
//func (as *AuthService) CheckUserVisitToken(ctx context.Context, visitToken string) bool {
//	accessToken, err := as.r.GetUserVisitCacheInfo(ctx, visitToken)
//	if err != nil {
//		return false
//	}
//	if len(accessToken) == 0 {
//		return false
//	}
//	return true
//}

//func (as *AuthService) RemoveUserVisitCacheInfo(ctx context.Context, visitToken string) (err error) {
//	if len(visitToken) > 0 {
//		return as.r.RemoveUserVisitCacheInfo(ctx, visitToken)
//	}
//	return nil
//}

//func (as *AuthService) AddUserTokenMapping(ctx context.Context, userID, accessToken string) (err error) {
//	return as.r.AddUserTokenMapping(ctx, userID, accessToken)
//}

//func (as *AuthService) RemoveUserAllTokens(ctx context.Context, userID string) {
//	as.r.RemoveUserTokens(ctx, userID, "")
//}

//func (as *AuthService) RemoveTokensExceptCurrentUser(ctx context.Context, userID string, accessToken string) {
//	as.r.RemoveUserTokens(ctx, userID, accessToken)
//}

//Admin
//
//func (as *AuthService) GetAdminUserCacheInfo(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error) {
//	return as.r.GetAdminUserCacheInfo(ctx, accessToken)
//}
//
//func (as *AuthService) SetAdminUserCacheInfo(ctx context.Context, accessToken string, userInfo *entity.UserCacheInfo) (err error) {
//	err = as.r.SetAdminUserCacheInfo(ctx, accessToken, userInfo)
//	return err
//}
//
//func (as *AuthService) RemoveAdminUserCacheInfo(ctx context.Context, accessToken string) (err error) {
//	return as.r.RemoveAdminUserCacheInfo(ctx, accessToken)
//}
