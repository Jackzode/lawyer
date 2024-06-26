package service

import (
	"context"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity "github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	checker "github.com/lawyer/commons/utils/checker"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/repo"
	"strings"

	"github.com/Chain-Zhang/pinyin"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/random"
	"github.com/segmentfault/pacman/errors"
)

type UserRepo interface {
	AddUser(ctx context.Context, user *entity.User) (err error)
	IncreaseAnswerCount(ctx context.Context, userID string, amount int) (err error)
	IncreaseQuestionCount(ctx context.Context, userID string, amount int) (err error)
	UpdateQuestionCount(ctx context.Context, userID string, count int64) (err error)
	UpdateAnswerCount(ctx context.Context, userID string, count int) (err error)
	UpdateLastLoginDate(ctx context.Context, userID string) (err error)
	UpdateEmailStatus(ctx context.Context, userID string, emailStatus int) error
	UpdateNoticeStatus(ctx context.Context, userID string, noticeStatus int) error
	UpdateEmail(ctx context.Context, userID, email string) error
	UpdateLanguage(ctx context.Context, userID, language string) error
	UpdatePass(ctx context.Context, userID, pass string) error
	UpdateInfo(ctx context.Context, userInfo *entity.User) (err error)
	GetByUserID(ctx context.Context, userID string) (userInfo *entity.User, exist bool, err error)
	BatchGetByID(ctx context.Context, ids []string) ([]*entity.User, error)
	GetUserInfoByUsername(ctx context.Context, username string) (userInfo *entity.User, exist bool, err error)
	GetByUsernames(ctx context.Context, usernames []string) ([]*entity.User, error)
	GetUserInfoByEmailFromDB(ctx context.Context, email string) (userInfo *entity.User, exist bool, err error)
	GetUserCount(ctx context.Context) (count int64, err error)
	SearchUserListByName(ctx context.Context, name string, limit int) (userList []*entity.User, err error)
	UpdateEmailAndEmailStatus(ctx context.Context, userID, email string, mailStatus int) (err error)
}

// UserCommonServicer user service
type UserCommon struct {
}

func NewUserCommon() *UserCommon {
	return &UserCommon{}
}

func (us *UserCommon) GetUserBasicInfoByID(ctx context.Context, ID string) (
	userBasicInfo *schema.UserBasicInfo, exist bool, err error) {
	userInfo, exist, err := repo.UserRepo.GetByUserID(ctx, ID)
	if err != nil {
		return nil, exist, err
	}
	info := us.FormatUserBasicInfo(ctx, userInfo)
	info.Avatar = schema.FormatAvatar(userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	return info, exist, nil
}

func (us *UserCommon) GetUserBasicInfoByUserName(ctx context.Context, username string) (*schema.UserBasicInfo, bool, error) {
	userInfo, exist, err := repo.UserRepo.GetUserInfoByUsername(ctx, username)
	if err != nil {
		return nil, exist, err
	}
	info := us.FormatUserBasicInfo(ctx, userInfo)
	info.Avatar = schema.FormatAvatar(userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	return info, exist, nil
}

func (us *UserCommon) BatchGetUserBasicInfoByUserNames(ctx context.Context, usernames []string) (map[string]*schema.UserBasicInfo, error) {
	infomap := make(map[string]*schema.UserBasicInfo)
	list, err := repo.UserRepo.GetByUsernames(ctx, usernames)
	if err != nil {
		return infomap, err
	}
	avatarMapping := schema.FormatListAvatar(list)
	for _, user := range list {
		info := us.FormatUserBasicInfo(ctx, user)
		info.Avatar = avatarMapping[user.ID].GetURL()
		infomap[user.Username] = info
	}
	return infomap, nil
}

func (us *UserCommon) UpdateAnswerCount(ctx context.Context, userID string, num int) error {
	return repo.UserRepo.UpdateAnswerCount(ctx, userID, num)
}

func (us *UserCommon) UpdateQuestionCount(ctx context.Context, userID string, num int64) error {
	return repo.UserRepo.UpdateQuestionCount(ctx, userID, num)
}

func (us *UserCommon) BatchUserBasicInfoByID(ctx context.Context, userIDs []string) (map[string]*schema.UserBasicInfo, error) {
	userMap := make(map[string]*schema.UserBasicInfo)
	if len(userIDs) == 0 {
		return userMap, nil
	}
	userList, err := repo.UserRepo.BatchGetByID(ctx, userIDs)
	if err != nil {
		return userMap, err
	}
	avatarMapping := schema.FormatListAvatar(userList)
	for _, user := range userList {
		info := us.FormatUserBasicInfo(ctx, user)
		info.Avatar = avatarMapping[user.ID].GetURL()
		userMap[user.ID] = info
	}
	return userMap, nil
}

// FormatUserBasicInfo format user basic info
func (us *UserCommon) FormatUserBasicInfo(ctx context.Context, userInfo *entity.User) *schema.UserBasicInfo {
	userBasicInfo := &schema.UserBasicInfo{}
	userBasicInfo.ID = userInfo.ID
	userBasicInfo.Username = userInfo.Username
	userBasicInfo.Rank = userInfo.Rank
	userBasicInfo.DisplayName = userInfo.DisplayName
	userBasicInfo.Website = userInfo.Website
	userBasicInfo.Location = userInfo.Location
	userBasicInfo.Status = constant.ConvertUserStatus(userInfo.Status, userInfo.MailStatus)
	if userBasicInfo.Status == constant.UserDeleted {
		userBasicInfo.Avatar = ""
		userBasicInfo.DisplayName = "user" + converter.DeleteUserDisplay(userInfo.ID)
	}
	return userBasicInfo
}

// MakeUsername
// Generate a unique Username based on the displayName
func (us *UserCommon) MakeUsername(ctx context.Context, displayName string) (username string, err error) {
	// Chinese processing
	if has := checker.IsChinese(displayName); has {
		str, err := pinyin.New(displayName).Split("").Mode(pinyin.WithoutTone).Convert()
		if err != nil {
			return "", errors.BadRequest(reason.UsernameInvalid)
		} else {
			displayName = str
		}
	}

	username = strings.ReplaceAll(displayName, " ", "-")
	username = strings.ToLower(username)
	suffix := ""

	if checker.IsInvalidUsername(username) {
		return "", errors.BadRequest(reason.UsernameInvalid)
	}

	if checker.IsReservedUsername(username) {
		return "", errors.BadRequest(reason.UsernameInvalid)
	}

	for {
		_, has, err := repo.UserRepo.GetUserInfoByUsername(ctx, username+suffix)
		if err != nil {
			return "", err
		}
		if !has {
			break
		}
		suffix = random.UsernameSuffix()
	}
	return username + suffix, nil
}

func (us *UserCommon) CacheLoginUserInfo(ctx context.Context, userID string, userStatus, emailStatus int, externalID string) (
	accessToken string, userCacheInfo *entity.UserCacheInfo, err error) {

	roleID, err := UserRoleRelServicer.GetUserRole(ctx, userID)
	if err != nil {
		glog.Slog.Error(err)
	}

	userCacheInfo = &entity.UserCacheInfo{
		UserID:      userID,
		EmailStatus: emailStatus,
		UserStatus:  userStatus,
		RoleID:      roleID,
		ExternalID:  externalID,
	}

	accessToken, err = AuthServicer.SetUserCacheInfo(ctx, userCacheInfo)
	if err != nil {
		return "", nil, err
	}
	return accessToken, userCacheInfo, nil
}
