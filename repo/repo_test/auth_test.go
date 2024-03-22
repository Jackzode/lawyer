package repo_test

import (
	"context"
	"github.com/lawyer/commons/entity"
	"testing"

	"github.com/lawyer/repo/auth"
	"github.com/stretchr/testify/assert"
)

var (
	accessToken = "token"
	visitToken  = "visitToken"
	userID      = "1"
)

func Test_authRepo_SetUserCacheInfo(t *testing.T) {
	authRepo := auth.NewAuthRepo(testDataSource)

	err := authRepo.SetUserCacheInfo(context.TODO(), accessToken, visitToken, &entity.UserCacheInfo{UserID: userID})
	assert.NoError(t, err)

	cacheInfo, err := authRepo.GetUserCacheInfo(context.TODO(), accessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, cacheInfo.UserID)
}

func Test_authRepo_RemoveUserCacheInfo(t *testing.T) {
	authRepo := auth.NewAuthRepo(testDataSource)

	err := authRepo.SetUserCacheInfo(context.TODO(), accessToken, visitToken, &entity.UserCacheInfo{UserID: userID})
	assert.NoError(t, err)

	err = authRepo.RemoveUserCacheInfo(context.TODO(), accessToken)
	assert.NoError(t, err)

	userInfo, err := authRepo.GetUserCacheInfo(context.TODO(), accessToken)
	assert.NoError(t, err)
	assert.Nil(t, userInfo)
}

func Test_authRepo_SetUserStatus(t *testing.T) {
	authRepo := auth.NewAuthRepo(testDataSource)

	err := authRepo.SetUserStatus(context.TODO(), userID, &entity.UserCacheInfo{UserID: userID})
	assert.NoError(t, err)

	cacheInfo, err := authRepo.GetUserStatus(context.TODO(), userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, cacheInfo.UserID)
}
func Test_authRepo_RemoveUserStatus(t *testing.T) {
	authRepo := auth.NewAuthRepo(testDataSource)

	err := authRepo.SetUserStatus(context.TODO(), userID, &entity.UserCacheInfo{UserID: userID})
	assert.NoError(t, err)

	err = authRepo.RemoveUserStatus(context.TODO(), userID)
	assert.NoError(t, err)

	userInfo, err := authRepo.GetUserStatus(context.TODO(), userID)
	assert.NoError(t, err)
	assert.Nil(t, userInfo)
}

func Test_authRepo_SetAdminUserCacheInfo(t *testing.T) {
	authRepo := auth.NewAuthRepo(testDataSource)

	err := authRepo.SetAdminUserCacheInfo(context.TODO(), accessToken, &entity.UserCacheInfo{UserID: userID})
	assert.NoError(t, err)

	cacheInfo, err := authRepo.GetAdminUserCacheInfo(context.TODO(), accessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, cacheInfo.UserID)
}

func Test_authRepo_RemoveAdminUserCacheInfo(t *testing.T) {
	authRepo := auth.NewAuthRepo(testDataSource)

	err := authRepo.SetAdminUserCacheInfo(context.TODO(), accessToken, &entity.UserCacheInfo{UserID: userID})
	assert.NoError(t, err)

	err = authRepo.RemoveAdminUserCacheInfo(context.TODO(), accessToken)
	assert.NoError(t, err)

	userInfo, err := authRepo.GetAdminUserCacheInfo(context.TODO(), accessToken)
	assert.NoError(t, err)
	assert.Nil(t, userInfo)
}
