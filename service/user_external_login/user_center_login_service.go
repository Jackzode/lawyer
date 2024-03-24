package user_external_login

import (
	"context"
	"encoding/json"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity2 "github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/commons/utils/checker"
	"github.com/lawyer/service/activity"
	"github.com/lawyer/service/siteinfo_common"
	"github.com/lawyer/service/user_common"
	"time"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/random"
	"github.com/lawyer/plugin"
	"github.com/segmentfault/pacman/log"
)

// UserCenterLoginService user external login service
type UserCenterLoginService struct {
	userRepo              usercommon.UserRepo
	userExternalLoginRepo UserExternalLoginRepo
	userCommonService     *usercommon.UserCommon
	userActivity          activity.UserActiveActivityRepo
	siteInfoCommonService siteinfo_common.SiteInfoCommonService
}

// NewUserCenterLoginService new user external login service
func NewUserCenterLoginService(
	userRepo usercommon.UserRepo,
	userCommonService *usercommon.UserCommon,
	userExternalLoginRepo UserExternalLoginRepo,
	userActivity activity.UserActiveActivityRepo,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService,
) *UserCenterLoginService {
	return &UserCenterLoginService{
		userRepo:              userRepo,
		userCommonService:     userCommonService,
		userExternalLoginRepo: userExternalLoginRepo,
		userActivity:          userActivity,
		siteInfoCommonService: siteInfoCommonService,
	}
}

func (us *UserCenterLoginService) ExternalLogin(
	ctx context.Context, userCenter plugin.UserCenter, basicUserInfo *plugin.UserCenterBasicUserInfo) (
	resp *schema.UserExternalLoginResp, err error) {
	if len(basicUserInfo.ExternalID) == 0 {
		return &schema.UserExternalLoginResp{
			ErrTitle: translator.Tr(utils.GetLangByCtx(ctx), reason.UserAccessDenied),
			ErrMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.UserExternalLoginMissingUserID),
		}, nil
	}

	if len(basicUserInfo.Email) > 0 {
		// check whether site allow register or not
		siteInfo, err := us.siteInfoCommonService.GetSiteLogin(ctx)
		if err != nil {
			return nil, err
		}
		if !checker.EmailInAllowEmailDomain(basicUserInfo.Email, siteInfo.AllowEmailDomains) {
			log.Debugf("email domain not allowed: %s", basicUserInfo.Email)
			return &schema.UserExternalLoginResp{
				ErrTitle: translator.Tr(utils.GetLangByCtx(ctx), reason.UserAccessDenied),
				ErrMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.EmailIllegalDomainError),
			}, nil
		}
	}

	oldExternalLoginUserInfo, exist, err := us.userExternalLoginRepo.GetByExternalID(ctx,
		userCenter.Info().SlugName, basicUserInfo.ExternalID)
	if err != nil {
		return nil, err
	}
	if exist {
		// if user is already a member, login directly
		oldUserInfo, exist, err := us.userRepo.GetByUserID(ctx, oldExternalLoginUserInfo.UserID)
		if err != nil {
			return nil, err
		}
		if exist {
			// if user is deleted, do not allow login
			if oldUserInfo.Status == entity2.UserStatusDeleted {
				return &schema.UserExternalLoginResp{
					ErrTitle: translator.Tr(utils.GetLangByCtx(ctx), reason.UserAccessDenied),
					ErrMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.UserPageAccessDenied),
				}, nil
			}
			if err := us.userRepo.UpdateLastLoginDate(ctx, oldUserInfo.ID); err != nil {
				log.Errorf("update user last login date failed: %v", err)
			}
			accessToken, _, err := us.userCommonService.CacheLoginUserInfo(
				ctx, oldUserInfo.ID, oldUserInfo.MailStatus, oldUserInfo.Status, oldExternalLoginUserInfo.ExternalID)
			return &schema.UserExternalLoginResp{AccessToken: accessToken}, err
		}
	}

	// cache external user info, waiting for user enter email address.
	if userCenter.Description().MustAuthEmailEnabled && len(basicUserInfo.Email) == 0 {
		return &schema.UserExternalLoginResp{ErrMsg: "Requires authorized email to login"}, nil
	}

	oldUserInfo, err := us.registerNewUser(ctx, userCenter.Info().SlugName, basicUserInfo)
	if err != nil {
		return nil, err
	}

	us.activeUser(ctx, oldUserInfo)

	accessToken, _, err := us.userCommonService.CacheLoginUserInfo(
		ctx, oldUserInfo.ID, oldUserInfo.MailStatus, oldUserInfo.Status, oldExternalLoginUserInfo.ExternalID)
	return &schema.UserExternalLoginResp{AccessToken: accessToken}, err
}

func (us *UserCenterLoginService) registerNewUser(ctx context.Context, provider string,
	basicUserInfo *plugin.UserCenterBasicUserInfo) (userInfo *entity2.User, err error) {
	userInfo = &entity2.User{}
	userInfo.EMail = basicUserInfo.Email
	userInfo.DisplayName = basicUserInfo.DisplayName

	userInfo.Username, err = us.userCommonService.MakeUsername(ctx, basicUserInfo.Username)
	if err != nil {
		log.Error(err)
		userInfo.Username = random.Username()
	}

	if len(basicUserInfo.Avatar) > 0 {
		avatarInfo := &schema.AvatarInfo{
			Type:   constant.AvatarTypeCustom,
			Custom: basicUserInfo.Avatar,
		}
		avatar, _ := json.Marshal(avatarInfo)
		userInfo.Avatar = string(avatar)
	}

	userInfo.MailStatus = entity2.EmailStatusAvailable
	userInfo.Status = entity2.UserStatusAvailable
	userInfo.LastLoginDate = time.Now()
	userInfo.Bio = basicUserInfo.Bio
	userInfo.BioHTML = converter.Markdown2HTML(basicUserInfo.Bio)
	err = us.userRepo.AddUser(ctx, userInfo)
	if err != nil {
		return nil, err
	}

	metaInfo, _ := json.Marshal(basicUserInfo)
	newExternalUserInfo := &entity2.UserExternalLogin{
		UserID:     userInfo.ID,
		Provider:   provider,
		ExternalID: basicUserInfo.ExternalID,
		MetaInfo:   string(metaInfo),
	}
	err = us.userExternalLoginRepo.AddUserExternalLogin(ctx, newExternalUserInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (us *UserCenterLoginService) activeUser(ctx context.Context, oldUserInfo *entity2.User) {
	if err := us.userActivity.UserActive(ctx, oldUserInfo.ID); err != nil {
		log.Error(err)
	}
}

func (us *UserCenterLoginService) UserCenterUserSettings(ctx context.Context, userID string) (
	resp *schema.UserCenterUserSettingsResp, err error) {
	resp = &schema.UserCenterUserSettingsResp{}

	userCenter, ok := plugin.GetUserCenter()
	if !ok {
		return resp, nil
	}

	// get external login info
	externalLoginList, err := us.userExternalLoginRepo.GetUserExternalLoginList(ctx, userID)
	if err != nil {
		return nil, err
	}
	var externalInfo *entity2.UserExternalLogin
	for _, t := range externalLoginList {
		if t.Provider == userCenter.Info().SlugName {
			externalInfo = t
		}
	}
	if externalInfo == nil {
		return resp, nil
	}

	settings, err := userCenter.UserSettings(externalInfo.ExternalID)
	if err != nil {
		log.Error(err)
		return resp, nil
	}

	if len(settings.AccountSettingRedirectURL) > 0 {
		resp.AccountSettingAgent = schema.UserSettingAgent{
			Enabled:     true,
			RedirectURL: settings.AccountSettingRedirectURL,
		}
	}
	if len(settings.ProfileSettingRedirectURL) > 0 {
		resp.ProfileSettingAgent = schema.UserSettingAgent{
			Enabled:     true,
			RedirectURL: settings.ProfileSettingRedirectURL,
		}
	}
	return resp, nil
}

// UserCenterAdminFunctionAgent Check in the backend administration interface if the user-related functions
// are turned off due to turning on the User Center plugin.
func (us *UserCenterLoginService) UserCenterAdminFunctionAgent(ctx context.Context) (
	resp *schema.UserCenterAdminFunctionAgentResp, err error) {
	resp = &schema.UserCenterAdminFunctionAgentResp{
		AllowCreateUser:         true,
		AllowUpdateUserStatus:   true,
		AllowUpdateUserPassword: true,
		AllowUpdateUserRole:     true,
	}
	userCenter, ok := plugin.GetUserCenter()
	if !ok {
		return
	}
	desc := userCenter.Description()
	// If user status agent is enabled, admin can not update user status in answer.
	resp.AllowUpdateUserStatus = !desc.UserStatusAgentEnabled
	resp.AllowUpdateUserRole = !desc.UserRoleAgentEnabled

	// If original user system is enabled, admin can update user password and role in answer.
	resp.AllowUpdateUserPassword = desc.EnabledOriginalUserSystem
	resp.AllowCreateUser = desc.EnabledOriginalUserSystem
	return resp, nil
}

func (us *UserCenterLoginService) UserCenterPersonalBranding(ctx context.Context, username string) (
	resp *schema.UserCenterPersonalBranding, err error) {
	resp = &schema.UserCenterPersonalBranding{
		PersonalBranding: make([]*schema.PersonalBranding, 0),
	}
	userCenter, ok := plugin.GetUserCenter()
	if !ok {
		return
	}

	userInfo, exist, err := us.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if !exist {
		return resp, nil
	}

	// get external login info
	externalLoginList, err := us.userExternalLoginRepo.GetUserExternalLoginList(ctx, userInfo.ID)
	if err != nil {
		return nil, err
	}
	var externalInfo *entity2.UserExternalLogin
	for _, t := range externalLoginList {
		if t.Provider == userCenter.Info().SlugName {
			externalInfo = t
		}
	}
	if externalInfo == nil {
		return resp, nil
	}

	resp.Enabled = true
	branding := userCenter.PersonalBranding(externalInfo.ExternalID)

	for _, t := range branding {
		resp.PersonalBranding = append(resp.PersonalBranding, &schema.PersonalBranding{
			Icon:  t.Icon,
			Name:  t.Name,
			Label: t.Label,
			Url:   t.Url,
		})
	}
	return resp, nil
}
