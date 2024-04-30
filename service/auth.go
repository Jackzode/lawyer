package service

import (
	"context"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/repo"

	"github.com/lawyer/pkg/token"
)

// AuthRepo auth repository
type AuthRepo interface {
	GetUserInfoFromCache(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error)
	SetUserCacheInfo(ctx context.Context, accessToken, visitToken string, userInfo *entity.UserCacheInfo) error
	GetUserVisitCacheInfo(ctx context.Context, visitToken string) (accessToken string, err error)
	RemoveUserCacheInfo(ctx context.Context, accessToken string) (err error)
	RemoveUserVisitCacheInfo(ctx context.Context, visitToken string) (err error)
	SetUserStatus(ctx context.Context, userID string, userInfo *entity.UserCacheInfo) (err error)
	GetUserStatusFromCache(ctx context.Context, userID string) (userInfo *entity.UserCacheInfo, err error)
	RemoveUserStatus(ctx context.Context, userID string) (err error)
	GetAdminUserCacheInfo(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error)
	SetAdminUserCacheInfo(ctx context.Context, accessToken string, userInfo *entity.UserCacheInfo) error
	RemoveAdminUserCacheInfo(ctx context.Context, accessToken string) (err error)
	AddUserTokenMapping(ctx context.Context, userID, accessToken string) (err error)
	RemoveUserTokens(ctx context.Context, userID string, remainToken string)
}

// AuthServicer kit service
type AuthService struct {
}

// NewAuthService email service
func NewAuthService() *AuthService {
	return &AuthService{}
}

func (as *AuthService) GetUserCacheInfo(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error) {
	userCacheInfo, err := repo.AuthRepo.GetUserInfoFromCache(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if userCacheInfo == nil {
		return nil, nil
	}
	userStatusInfo, _ := repo.AuthRepo.GetUserStatusFromCache(ctx, userCacheInfo.UserID)
	if userStatusInfo != nil {
		userCacheInfo.UserStatus = userStatusInfo.UserStatus
		userCacheInfo.EmailStatus = userStatusInfo.EmailStatus
		userCacheInfo.RoleID = userStatusInfo.RoleID
		// update current user cache info
		err = repo.AuthRepo.SetUserCacheInfo(ctx, accessToken, userCacheInfo.VisitToken, userCacheInfo)
		if err != nil {
			return nil, err
		}
	}

	//todo  try to get user status from user center
	//uc, ok := plugin.GetUserCenter()
	//if ok && len(userCacheInfo.ExternalID) > 0 {
	//	if userStatus := uc.UserStatus(userCacheInfo.ExternalID); userStatus != plugin.UserStatusAvailable {
	//		userCacheInfo.UserStatus = int(userStatus)
	//	}
	//}
	return userCacheInfo, nil
}

// visit token 指向access token， acctoken指向userCacheInfo
func (as *AuthService) SetUserCacheInfo(ctx context.Context, userInfo *entity.UserCacheInfo) (
	accessToken string, visitToken string, err error) {
	accessToken = token.GenerateToken()
	visitToken = token.GenerateToken()
	err = repo.AuthRepo.SetUserCacheInfo(ctx, accessToken, visitToken, userInfo)
	if err != nil {
		return "", "", err
	}
	return accessToken, visitToken, err
}

func (as *AuthService) CheckUserVisitToken(ctx context.Context, visitToken string) bool {
	accessToken, err := repo.AuthRepo.GetUserVisitCacheInfo(ctx, visitToken)
	if err != nil {
		return false
	}
	if len(accessToken) == 0 {
		return false
	}
	return true
}

func (as *AuthService) SetUserStatus(ctx context.Context, userInfo *entity.UserCacheInfo) (err error) {
	return repo.AuthRepo.SetUserStatus(ctx, userInfo.UserID, userInfo)
}

func (as *AuthService) RemoveUserCacheInfo(ctx context.Context, accessToken string) (err error) {
	return repo.AuthRepo.RemoveUserCacheInfo(ctx, accessToken)
}

func (as *AuthService) RemoveUserVisitCacheInfo(ctx context.Context, visitToken string) (err error) {
	if len(visitToken) > 0 {
		return repo.AuthRepo.RemoveUserVisitCacheInfo(ctx, visitToken)
	}
	return nil
}

// AddUserTokenMapping add user token mapping
func (as *AuthService) AddUserTokenMapping(ctx context.Context, userID, accessToken string) (err error) {
	return repo.AuthRepo.AddUserTokenMapping(ctx, userID, accessToken)
}

// RemoveUserAllTokens Log out all users under this user id
func (as *AuthService) RemoveUserAllTokens(ctx context.Context, userID string) {
	repo.AuthRepo.RemoveUserTokens(ctx, userID, "")
}

// RemoveTokensExceptCurrentUser remove all tokens except the current user
func (as *AuthService) RemoveTokensExceptCurrentUser(ctx context.Context, userID string, accessToken string) {
	repo.AuthRepo.RemoveUserTokens(ctx, userID, accessToken)
}

//Admin

func (as *AuthService) GetAdminUserCacheInfo(ctx context.Context, accessToken string) (userInfo *entity.UserCacheInfo, err error) {
	return repo.AuthRepo.GetAdminUserCacheInfo(ctx, accessToken)
}

func (as *AuthService) SetAdminUserCacheInfo(ctx context.Context, accessToken string, userInfo *entity.UserCacheInfo) (err error) {
	err = repo.AuthRepo.SetAdminUserCacheInfo(ctx, accessToken, userInfo)
	return err
}

func (as *AuthService) RemoveAdminUserCacheInfo(ctx context.Context, accessToken string) (err error) {
	return repo.AuthRepo.RemoveAdminUserCacheInfo(ctx, accessToken)
}
