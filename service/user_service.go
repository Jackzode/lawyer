package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity "github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/utils"
	checker "github.com/lawyer/commons/utils/checker"
	"github.com/lawyer/repo"
	"github.com/lawyer/repoCommon"
	"time"

	"github.com/google/uuid"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/plugin"
	"github.com/segmentfault/pacman/errors"
	"golang.org/x/crypto/bcrypt"
)

// UserServicer user service
type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

// GetUserInfoByUserID get user info by user id
func (us *UserService) GetUserInfoByUserID(ctx context.Context, token, userID string) (
	userInfo *entity.User, err error) {

	userInfo, exist, err := repo.UserRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	if userInfo.Status == entity.UserStatusDeleted {
		return nil, errors.Unauthorized(reason.UnauthorizedError)
	}
	return userInfo, nil

}

func (us *UserService) GetOtherUserInfoByUsername(ctx context.Context, username string) (
	resp *schema.GetOtherUserInfoByUsernameResp, err error) {
	userInfo, exist, err := repo.UserRepo.GetUserInfoByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.UserNotFound)
	}
	resp = &schema.GetOtherUserInfoByUsernameResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.Avatar = SiteInfoCommonServicer.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	return resp, nil
}

/*
检查邮箱状态是否正常，也就是账号状态
对比密码是否正确
然后更新最近登录时间
*/
func (us *UserService) EmailLogin(ctx context.Context, req *schema.UserEmailLoginReq) (resp *schema.UserLoginResp, err error) {

	userInfo, exist, err := repo.UserRepo.GetUserInfoByEmailFromDB(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if !exist || userInfo.Status == entity.UserStatusDeleted {
		return nil, errors.BadRequest(reason.EmailOrPasswordWrong)
	}
	if !us.verifyPassword(ctx, req.Pass, userInfo.Pass) {
		return nil, errors.BadRequest(reason.EmailOrPasswordWrong)
	}
	//更新最近登陆时间
	err = repo.UserRepo.UpdateLastLoginDate(ctx, userInfo.ID)
	if err != nil {
		glog.Slog.Errorf("update last glog.Slogin data failed, err: %v", err)
	}
	//查db获取用户角色id
	roleID, err := UserRoleRelServicer.GetUserRole(ctx, userInfo.ID)
	if err != nil {
		glog.Slog.Error(err)
	}

	resp = &schema.UserLoginResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.RoleID = roleID
	//生成用户头像
	resp.Avatar = SiteInfoCommonServicer.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	userCacheInfo := &entity.UserCacheInfo{
		UserID:      userInfo.ID,
		EmailStatus: userInfo.MailStatus,
		UserStatus:  userInfo.Status,
		RoleID:      roleID,
		//ExternalID:  externalID,
	}
	resp.AccessToken, err = AuthServicer.SetUserCacheInfo(ctx, userCacheInfo)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// RetrievePassWord
// 通过email获取用户信息,然后给这个邮箱地址发个邮件
func (us *UserService) RetrievePassWord(ctx context.Context, req *schema.UserRetrievePassWordRequest) error {

	userInfo, has, err := repo.UserRepo.GetUserInfoByEmailFromDB(ctx, req.Email)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	// send email
	data := &schema.EmailCodeContent{
		Email:  req.Email,
		UserID: userInfo.ID,
	}
	code := uuid.NewString()
	verifyEmailURL := fmt.Sprintf("%s/users/password-reset?code=%s", us.getSiteUrl(ctx), code)
	title, body, err := EmailServicer.PassResetTemplate(ctx, verifyEmailURL)
	if err != nil {
		return err
	}
	go EmailServicer.SendAndSaveCode(ctx, req.Email, title, body, code, data.ToJSONString())
	return nil
}

// UpdatePasswordWhenForgot update user password when user forgot password
func (us *UserService) UpdatePasswordWhenForgot(ctx context.Context, req *schema.UserRePassWordRequest) (err error) {
	data := &schema.EmailCodeContent{}
	//这个content是通过code从缓存里拿到的，里面包含的是用户信息，应该是给用户发修改密码邮件前存的
	err = data.FromJSONString(req.Content)
	if err != nil {
		return errors.BadRequest(reason.EmailVerifyURLExpired)
	}
	//从db中查询用户信息
	userInfo, exist, err := repo.UserRepo.GetUserInfoByEmailFromDB(ctx, data.Email)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}
	//加密新密码
	newPass, err := utils.EncryptPassword(req.Pass)
	if err != nil {
		return err
	}
	//更新
	err = repo.UserRepo.UpdatePass(ctx, userInfo.ID, newPass)
	if err != nil {
		return err
	}
	// When the user changes the password, all the current user's tokens are invalid.
	//todo 确实需要删除旧的token，那么服务端如何删除这个token？这个也涉及到如何refresh token
	//AuthServicer.RemoveUserAllTokens(ctx, userInfo.ID)
	return nil
}

func (us *UserService) UserPassWordVerification(ctx context.Context, uid, oldPass string) (bool, error) {
	userInfo, has, err := repo.UserRepo.GetByUserID(ctx, uid)
	if err != nil {
		return false, err
	}
	if !has {
		return false, errors.BadRequest(reason.UserNotFound)
	}
	isPass := us.verifyPassword(ctx, oldPass, userInfo.Pass)
	if !isPass {
		return false, nil
	}

	return true, nil
}

// UserModifyPassword user modify password
func (us *UserService) UserModifyPassword(ctx context.Context, req *schema.UserModifyPasswordReq) error {
	enpass, err := utils.EncryptPassword(req.Pass)
	if err != nil {
		return err
	}
	userInfo, exist, err := repo.UserRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}
	//再次验证老密码是否正确
	isPass := us.verifyPassword(ctx, req.OldPass, userInfo.Pass)
	if !isPass {
		return errors.BadRequest(reason.OldPasswordVerificationFailed)
	}
	//更新数据库密码
	err = repo.UserRepo.UpdatePass(ctx, userInfo.ID, enpass)
	if err != nil {
		return err
	}

	//AuthServicer.RemoveTokensExceptCurrentUser(ctx, userInfo.ID, req.AccessToken)
	return nil
}

// UpdateInfo update user info
func (us *UserService) UpdateInfo(ctx context.Context, req *schema.UpdateInfoRequest) (err error) {
	//siteUsers, err := SiteInfoCommonServicer.GetSiteUsers(ctx)
	//if err != nil {
	//	return nil, err
	//}

	if len(req.Username) > 0 {

		// admin can use reserved username
		if checker.IsInvalidUsername(req.Username) || checker.IsReservedUsername(req.Username) || checker.IsUsersIgnorePath(req.Username) {
			return errors.BadRequest(reason.UsernameInvalid)
		}

		userInfo, exist, err := repo.UserRepo.GetUserInfoByUsername(ctx, req.Username)
		if err != nil {
			return err
		}
		if exist && userInfo.ID != req.UserID {
			return errors.BadRequest(reason.UsernameDuplicate)
		}
	}

	oldUserInfo, exist, err := repo.UserRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}

	cond := us.formatUserInfoForUpdateInfo(oldUserInfo, req)
	err = repo.UserRepo.UpdateInfo(ctx, cond)
	return err
}

func (us *UserService) formatUserInfoForUpdateInfo(
	oldUserInfo *entity.User, req *schema.UpdateInfoRequest) *entity.User {
	avatar, _ := json.Marshal(req.Avatar)

	userInfo := &entity.User{}
	userInfo.DisplayName = oldUserInfo.DisplayName
	userInfo.Username = oldUserInfo.Username
	userInfo.Avatar = oldUserInfo.Avatar
	userInfo.Bio = oldUserInfo.Bio
	userInfo.BioHTML = oldUserInfo.BioHTML
	userInfo.Website = oldUserInfo.Website
	userInfo.Location = oldUserInfo.Location
	userInfo.ID = req.UserID

	if len(req.DisplayName) > 0 {
		userInfo.DisplayName = req.DisplayName
	}
	if len(req.Username) > 0 {
		userInfo.Username = req.Username
	}
	if len(avatar) > 0 {
		userInfo.Avatar = string(avatar)
	}
	userInfo.Bio = req.Bio
	userInfo.BioHTML = req.BioHTML
	userInfo.Website = req.Website
	userInfo.Location = req.Location
	return userInfo
}

// UserUpdateInterface update user interface
func (us *UserService) UserUpdateInterface(ctx context.Context, req *schema.UpdateUserInterfaceRequest) (err error) {

	err = repo.UserRepo.UpdateLanguage(ctx, req.UserId, req.Language)
	if err != nil {
		return
	}
	return nil
}

// UserRegisterByEmail user register
func (us *UserService) UserRegisterByEmail(ctx context.Context, registerUserInfo *schema.UserRegisterReq) (
	resp *schema.UserLoginResp, err error) {
	//先查一下数据库是否有这个邮箱地址，有则是重复注册
	_, has, err := repo.UserRepo.GetUserInfoByEmailFromDB(ctx, registerUserInfo.Email)
	if err != nil {
		return nil, err
	}
	//邮箱重复了
	if has {
		return nil, errors.BadRequest(reason.EmailDuplicate)
	}

	userInfo := &entity.User{}
	userInfo.EMail = registerUserInfo.Email
	userInfo.DisplayName = registerUserInfo.Name
	userInfo.Pass, err = utils.EncryptPassword(registerUserInfo.Pass)
	if err != nil {
		return nil, err
	}
	userInfo.Username, err = UserCommonServicer.MakeUsername(ctx, registerUserInfo.Name)
	if err != nil {
		return nil, err
	}
	userInfo.IPInfo = registerUserInfo.IP
	userInfo.MailStatus = entity.EmailStatusToBeVerified
	userInfo.Status = entity.UserStatusAvailable
	userInfo.LastLoginDate = time.Now()
	//userInfo.ID是插入mysql生成的
	err = repo.UserRepo.AddUser(ctx, userInfo)
	if err != nil {
		return nil, err
	}
	//todo
	if err = UserNotificationConfigService.SetDefaultUserNotificationConfig(ctx, []string{userInfo.ID}); err != nil {
		glog.Klog.Error("set default user notification config failed, err: " + err.Error())
	}

	// send email
	data := &schema.EmailCodeContent{
		Email:  registerUserInfo.Email,
		UserID: userInfo.ID,
	}
	code := uuid.NewString()
	verifyEmailURL := fmt.Sprintf("%s/users/account-activation?code=%s", us.getSiteUrl(ctx), code)
	title, body, err := EmailServicer.RegisterTemplate(ctx, verifyEmailURL)
	if err != nil {
		return nil, err
	}
	go EmailServicer.SendAndSaveCode(ctx, userInfo.EMail, title, body, code, data.ToJSONString())
	//新注册用户不存在role id，默认为1
	roleID, err := UserRoleRelServicer.GetUserRole(ctx, userInfo.ID)
	if err != nil {
		glog.Klog.Error(err.Error())
	}

	// return user info and token
	resp = &schema.UserLoginResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.RoleID = roleID
	//todo
	resp.Avatar = SiteInfoCommonServicer.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	userCacheInfo := &entity.UserCacheInfo{
		UserID:      userInfo.ID,
		EmailStatus: userInfo.MailStatus,
		UserStatus:  userInfo.Status,
		RoleID:      roleID,
		UserName:    userInfo.Username,
	}
	//acctoken指向userCacheInfo
	resp.AccessToken, err = AuthServicer.SetUserCacheInfo(ctx, userCacheInfo)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (us *UserService) UserVerifyEmailSend(ctx context.Context, userID string) error {
	userInfo, has, err := repo.UserRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if !has {
		return errors.BadRequest(reason.UserNotFound)
	}

	data := &schema.EmailCodeContent{
		Email:  userInfo.EMail,
		UserID: userInfo.ID,
	}
	code := uuid.NewString()
	verifyEmailURL := fmt.Sprintf("%s/users/account-activation?code=%s", us.getSiteUrl(ctx), code)
	title, body, err := EmailServicer.RegisterTemplate(ctx, verifyEmailURL)
	if err != nil {
		return err
	}
	go EmailServicer.SendAndSaveCode(ctx, userInfo.EMail, title, body, code, data.ToJSONString())
	return nil
}

func (us *UserService) UserVerifyEmail(ctx context.Context, req *schema.UserVerifyEmailReq) (resp *schema.UserLoginResp, err error) {
	data := &schema.EmailCodeContent{}
	err = data.FromJSONString(req.Content)
	if err != nil {
		return nil, errors.BadRequest(reason.EmailVerifyURLExpired)
	}
	//根据content里的email和uid，查db获取用户的全部信息
	userInfo, has, err := repo.UserRepo.GetUserInfoByEmailFromDB(ctx, data.Email)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	if userInfo.MailStatus == entity.EmailStatusToBeVerified {
		userInfo.MailStatus = entity.EmailStatusAvailable
		//更新用户的邮箱状态为激活状态
		err = repo.UserRepo.UpdateEmailStatus(ctx, userInfo.ID, userInfo.MailStatus)
		if err != nil {
			return nil, err
		}
	}
	//记录用户activity事件，修改用户rank排名，在一个事务里做的
	if err = repo.UserActiveActivityRepo.UserActive(ctx, userInfo.ID); err != nil {
		glog.Klog.Error(err.Error())
	}

	// In the case of three-party login, the associated users are bound
	//if len(data.BindingKey) > 0 {
	//	err = UserExternalLoginServicer.ExternalLoginBindingUser(ctx, data.BindingKey, userInfo)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	//这里为啥又缓存了一下用户信息？在注册的时候已经保存了一次
	//但是之前缓存信息里的email status过期了，需要更新，这里如何做也是个问题 todo
	accessToken, userCacheInfo, err := UserCommonServicer.CacheLoginUserInfo(ctx, userInfo.ID, userInfo.MailStatus, userInfo.Status, "")
	if err != nil {
		return nil, err
	}

	resp = &schema.UserLoginResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.Avatar = SiteInfoCommonServicer.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	resp.AccessToken = accessToken
	// User verified email will update user email status. So user status cache should be updated.
	if err = AuthServicer.SetUserCacheInfoByUid(ctx, userCacheInfo); err != nil {
		return nil, err
	}
	return resp, nil
}

// verifyPassword
// Compare whether the password is correct
func (us *UserService) verifyPassword(ctx context.Context, loginPass, userPass string) bool {
	if len(loginPass) == 0 && len(userPass) == 0 {
		return true
	}
	err := bcrypt.CompareHashAndPassword([]byte(userPass), []byte(loginPass))
	return err == nil
}

// encryptPassword
// The password does irreversible encryption.

func (us *UserService) UserChangeEmailSendCode(ctx context.Context, req *schema.UserChangeEmailSendCodeReq) (
	resp []*validator.FormErrorField, err error) {
	//根据uid查询用户信息
	userInfo, exist, err := repo.UserRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	//校对邮箱状态和密码
	// If user's email already verified, then must verify password first.
	if userInfo.MailStatus == entity.EmailStatusAvailable && !us.verifyPassword(ctx, req.Pass, userInfo.Pass) {
		resp = append(resp, &validator.FormErrorField{
			ErrorField: "pass",
			ErrorMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.OldPasswordVerificationFailed),
		})
		return resp, errors.BadRequest(reason.OldPasswordVerificationFailed)
	}
	//确认下是否是重复的邮箱
	_, exist, err = repo.UserRepo.GetUserInfoByEmailFromDB(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exist {
		resp = append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "e_mail",
			ErrorMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.EmailDuplicate),
		})
		return resp, errors.BadRequest(reason.EmailDuplicate)
	}

	data := &schema.EmailCodeContent{
		Email:  req.Email,
		UserID: req.UserID,
	}
	code := uuid.NewString()
	var title, body string
	verifyEmailURL := fmt.Sprintf("%s/users/confirm-new-email?code=%s", us.getSiteUrl(ctx), code)
	if userInfo.MailStatus == entity.EmailStatusToBeVerified {
		title, body, err = EmailServicer.RegisterTemplate(ctx, verifyEmailURL)
	} else {
		title, body, err = EmailServicer.ChangeEmailTemplate(ctx, verifyEmailURL)
	}
	if err != nil {
		return nil, err
	}
	glog.Slog.Infof("send email confirmation %s", verifyEmailURL)
	//给新邮箱发送验证码
	go EmailServicer.SendAndSaveCode(ctx, req.Email, title, body, code, data.ToJSONString())
	return nil, nil
}

// UserChangeEmailVerify user change email verify code
func (us *UserService) UserChangeEmailVerify(ctx context.Context, content string) (resp *schema.UserLoginResp, err error) {
	data := &schema.EmailCodeContent{}
	err = data.FromJSONString(content)
	if err != nil {
		return nil, errors.BadRequest(reason.EmailVerifyURLExpired)
	}

	_, exist, err := repo.UserRepo.GetUserInfoByEmailFromDB(ctx, data.Email)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.BadRequest(reason.EmailDuplicate)
	}

	userInfo, exist, err := repo.UserRepo.GetByUserID(ctx, data.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	//更新db中的邮箱
	err = repo.UserRepo.UpdateEmail(ctx, data.UserID, data.Email)
	if err != nil {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	err = repo.UserRepo.UpdateEmailStatus(ctx, data.UserID, entity.EmailStatusAvailable)
	if err != nil {
		return nil, err
	}

	roleID, err := UserRoleRelServicer.GetUserRole(ctx, userInfo.ID)
	if err != nil {
		glog.Slog.Error(err)
	}

	resp = &schema.UserLoginResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.Avatar = SiteInfoCommonServicer.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	userCacheInfo := &entity.UserCacheInfo{
		UserID:      userInfo.ID,
		EmailStatus: entity.EmailStatusAvailable,
		UserStatus:  userInfo.Status,
		RoleID:      roleID,
	}
	resp.AccessToken, err = AuthServicer.SetUserCacheInfo(ctx, userCacheInfo)
	if err != nil {
		return nil, err
	}
	// User verified email will update user email status. So user status cache should be updated.
	if err = AuthServicer.SetUserCacheInfoByUid(ctx, userCacheInfo); err != nil {
		return nil, err
	}
	resp.RoleID = userCacheInfo.RoleID
	return resp, nil
}

// getSiteUrl get site url todo 改成读配置文件
func (us *UserService) getSiteUrl(ctx context.Context) string {
	//siteGeneral, err := SiteInfoCommonServicer.GetSiteGeneral(ctx)
	//if err != nil {
	//	glog.Slog.Errorf("get site general failed: %s", err)
	//	return ""
	//}
	//return siteGeneral.SiteUrl
	return ""
}

// UserRanking get user ranking
func (us *UserService) UserRanking(ctx context.Context) (resp *schema.UserRankingResp, err error) {
	limit := 20
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)
	userIDs, userIDExist := make([]string, 0), make(map[string]bool, 0)

	// get most reputation users
	rankStat, rankStatUserIDs, err := us.getActivityUserRankStat(ctx, startTime, endTime, limit, userIDExist)
	if err != nil {
		return nil, err
	}
	userIDs = append(userIDs, rankStatUserIDs...)

	// get most vote users
	voteStat, voteStatUserIDs, err := us.getActivityUserVoteStat(ctx, startTime, endTime, limit, userIDExist)
	if err != nil {
		return nil, err
	}
	userIDs = append(userIDs, voteStatUserIDs...)

	// get all staff members
	userRoleRels, staffUserIDs, err := us.getStaff(ctx, userIDExist)
	if err != nil {
		return nil, err
	}
	userIDs = append(userIDs, staffUserIDs...)

	// get user information
	userInfoMapping, err := us.getUserInfoMapping(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	return us.warpStatRankingResp(userInfoMapping, rankStat, voteStat, userRoleRels), nil
}

// UserUnsubscribeNotification user unsubscribe email notification
func (us *UserService) UserUnsubscribeNotification(
	ctx context.Context, req *schema.UserUnsubscribeNotificationReq) (err error) {
	data := &schema.EmailCodeContent{}
	err = data.FromJSONString(req.Content)
	if err != nil || len(data.UserID) == 0 {
		return errors.BadRequest(reason.EmailVerifyURLExpired)
	}

	for _, source := range data.NotificationSources {
		notificationConfig, exist, err := repo.UserNotificationConfigRepo.GetByUserIDAndSource(
			ctx, data.UserID, source)
		if err != nil {
			return err
		}
		if !exist {
			continue
		}
		channels := schema.NewNotificationChannelsFormJson(notificationConfig.Channels)
		// unsubscribe email notification
		for _, channel := range channels {
			if channel.Key == constant.EmailChannel {
				channel.Enable = false
			}
		}
		notificationConfig.Channels = channels.ToJsonString()
		if err = repo.UserNotificationConfigRepo.Save(ctx, notificationConfig); err != nil {
			return err
		}
	}
	return nil
}

func (us *UserService) getActivityUserRankStat(ctx context.Context, startTime, endTime time.Time, limit int,
	userIDExist map[string]bool) (rankStat []*entity.ActivityUserRankStat, userIDs []string, err error) {
	//if plugin.RankAgentEnabled() {
	//	return make([]*entity.ActivityUserRankStat, 0), make([]string, 0), nil
	//}
	rankStat, err = repoCommon.NewActivityRepo().GetUsersWhoHasGainedTheMostReputation(ctx, startTime, endTime, limit)
	if err != nil {
		return nil, nil, err
	}
	for _, stat := range rankStat {
		if stat.Rank <= 0 {
			continue
		}
		if userIDExist[stat.UserID] {
			continue
		}
		userIDs = append(userIDs, stat.UserID)
		userIDExist[stat.UserID] = true
	}
	return rankStat, userIDs, nil
}

func (us *UserService) getActivityUserVoteStat(ctx context.Context, startTime, endTime time.Time, limit int,
	userIDExist map[string]bool) (voteStat []*entity.ActivityUserVoteStat, userIDs []string, err error) {
	if plugin.RankAgentEnabled() {
		return make([]*entity.ActivityUserVoteStat, 0), make([]string, 0), nil
	}
	voteStat, err = repoCommon.NewActivityRepo().GetUsersWhoHasVoteMost(ctx, startTime, endTime, limit)
	if err != nil {
		return nil, nil, err
	}
	for _, stat := range voteStat {
		if stat.VoteCount <= 0 {
			continue
		}
		if userIDExist[stat.UserID] {
			continue
		}
		userIDs = append(userIDs, stat.UserID)
		userIDExist[stat.UserID] = true
	}
	return voteStat, userIDs, nil
}

func (us *UserService) getStaff(ctx context.Context, userIDExist map[string]bool) (
	userRoleRels []*entity.UserRoleRel, userIDs []string, err error) {
	userRoleRels, err = UserRoleRelServicer.GetUserByRoleID(ctx, []int{RoleAdminID, RoleModeratorID})
	if err != nil {
		return nil, nil, err
	}
	for _, rel := range userRoleRels {
		if userIDExist[rel.UserID] {
			continue
		}
		userIDs = append(userIDs, rel.UserID)
		userIDExist[rel.UserID] = true
	}
	return userRoleRels, userIDs, nil
}

func (us *UserService) getUserInfoMapping(ctx context.Context, userIDs []string) (
	userInfoMapping map[string]*entity.User, err error) {
	userInfoMapping = make(map[string]*entity.User, 0)
	if len(userIDs) == 0 {
		return userInfoMapping, nil
	}
	userInfoList, err := repo.UserRepo.BatchGetByID(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	avatarMapping := SiteInfoCommonServicer.FormatListAvatar(ctx, userInfoList)
	for _, user := range userInfoList {
		user.Avatar = avatarMapping[user.ID].GetURL()
		userInfoMapping[user.ID] = user
	}
	return userInfoMapping, nil
}

func (us *UserService) SearchUserListByName(ctx context.Context, req *schema.GetOtherUserInfoByUsernameReq) (
	resp []*schema.UserBasicInfo, err error) {
	resp = make([]*schema.UserBasicInfo, 0)
	if len(req.Username) == 0 {
		return resp, nil
	}
	//根据username或者display name查db
	userList, err := repo.UserRepo.SearchUserListByName(ctx, req.Username, 5)
	if err != nil {
		return resp, err
	}
	avatarMapping := SiteInfoCommonServicer.FormatListAvatar(ctx, userList)
	for _, u := range userList {
		if req.UserID == u.ID {
			//搜到了自己，就跳过
			continue
		}
		basicInfo := UserCommonServicer.FormatUserBasicInfo(ctx, u)
		basicInfo.Avatar = avatarMapping[u.ID].GetURL()
		resp = append(resp, basicInfo)
	}
	return resp, nil
}

func (us *UserService) warpStatRankingResp(
	userInfoMapping map[string]*entity.User,
	rankStat []*entity.ActivityUserRankStat,
	voteStat []*entity.ActivityUserVoteStat,
	userRoleRels []*entity.UserRoleRel) (resp *schema.UserRankingResp) {
	resp = &schema.UserRankingResp{
		UsersWithTheMostReputation: make([]*schema.UserRankingSimpleInfo, 0),
		UsersWithTheMostVote:       make([]*schema.UserRankingSimpleInfo, 0),
		Staffs:                     make([]*schema.UserRankingSimpleInfo, 0),
	}
	for _, stat := range rankStat {
		if stat.Rank <= 0 {
			continue
		}
		if userInfo := userInfoMapping[stat.UserID]; userInfo != nil && userInfo.Status != entity.UserStatusDeleted {
			resp.UsersWithTheMostReputation = append(resp.UsersWithTheMostReputation, &schema.UserRankingSimpleInfo{
				Username:    userInfo.Username,
				Rank:        stat.Rank,
				DisplayName: userInfo.DisplayName,
				Avatar:      userInfo.Avatar,
			})
		}
	}
	for _, stat := range voteStat {
		if stat.VoteCount <= 0 {
			continue
		}
		if userInfo := userInfoMapping[stat.UserID]; userInfo != nil && userInfo.Status != entity.UserStatusDeleted {
			resp.UsersWithTheMostVote = append(resp.UsersWithTheMostVote, &schema.UserRankingSimpleInfo{
				Username:    userInfo.Username,
				VoteCount:   stat.VoteCount,
				DisplayName: userInfo.DisplayName,
				Avatar:      userInfo.Avatar,
			})
		}
	}
	for _, rel := range userRoleRels {
		if userInfo := userInfoMapping[rel.UserID]; userInfo != nil && userInfo.Status != entity.UserStatusDeleted {
			resp.Staffs = append(resp.Staffs, &schema.UserRankingSimpleInfo{
				Username:    userInfo.Username,
				Rank:        userInfo.Rank,
				DisplayName: userInfo.DisplayName,
				Avatar:      userInfo.Avatar,
			})
		}
	}
	return resp
}
