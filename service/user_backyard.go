package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/site"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/repo"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"github.com/segmentfault/pacman/errors"
	"golang.org/x/crypto/bcrypt"
)

// UserAdminRepo user repository
type UserAdminRepo interface {
	UpdateUserStatus(ctx context.Context, userID string, userStatus, mailStatus int, email string) (err error)
	GetUserInfo(ctx context.Context, userID string) (user *entity.User, exist bool, err error)
	GetUserInfoByEmail(ctx context.Context, email string) (user *entity.User, exist bool, err error)
	GetUserPage(ctx context.Context, page, pageSize int, user *entity.User,
		usernameOrDisplayName string, isStaff bool) (users []*entity.User, total int64, err error)
	AddUser(ctx context.Context, user *entity.User) (err error)
	AddUsers(ctx context.Context, users []*entity.User) (err error)
	UpdateUserPassword(ctx context.Context, userID string, password string) (err error)
}

// UserAdminServicer user service
type UserAdminService struct {
}

// NewUserAdminService new user admin service
func NewUserAdminService() *UserAdminService {
	return &UserAdminService{}
}

// UpdateUserStatus update user
func (us *UserAdminService) UpdateUserStatus(ctx context.Context, req *schema.UpdateUserStatusReq) (err error) {
	// Admin cannot modify their status
	if req.UserID == req.LoginUserID {
		return errors.BadRequest(reason.AdminCannotModifySelfStatus)
	}
	userInfo, exist, err := repo.UserAdminRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}
	// if user status is deleted
	if userInfo.Status == entity.UserStatusDeleted {
		return nil
	}

	if req.IsInactive() {
		userInfo.MailStatus = entity.EmailStatusToBeVerified
	}
	if req.IsDeleted() {
		userInfo.Status = entity.UserStatusDeleted
		userInfo.EMail = fmt.Sprintf("%s.%d", userInfo.EMail, time.Now().Unix())
	}
	if req.IsSuspended() {
		userInfo.Status = entity.UserStatusSuspended
	}
	if req.IsNormal() {
		userInfo.Status = entity.UserStatusAvailable
		userInfo.MailStatus = entity.EmailStatusAvailable
	}

	err = repo.UserAdminRepo.UpdateUserStatus(ctx, userInfo.ID, userInfo.Status, userInfo.MailStatus, userInfo.EMail)
	if err != nil {
		return err
	}

	// remove all content that user created, such as question, answer, comment, etc.
	if req.RemoveAllContent {
		us.removeAllUserCreatedContent(ctx, userInfo.ID)
	}

	// if user reputation is zero means this user is inactive, so try to activate this user.
	if req.IsNormal() && userInfo.Rank == 0 {
		return repo.UserActiveActivityRepo.UserActive(ctx, userInfo.ID)
	}
	return nil
}

// removeAllUserCreatedContent remove all user created content
func (us *UserAdminService) removeAllUserCreatedContent(ctx context.Context, userID string) {
	if err := repo.QuestionRepo.RemoveAllUserQuestion(ctx, userID); err != nil {
		glog.Slog.Errorf("remove all user question error: %v", err)
	}
	if err := repo.AnswerRepo.RemoveAllUserAnswer(ctx, userID); err != nil {
		glog.Slog.Errorf("remove all user answer error: %v", err)
	}
	if err := repo.CommentCommonRepo.RemoveAllUserComment(ctx, userID); err != nil {
		glog.Slog.Errorf("remove all user comment error: %v", err)
	}
}

// UpdateUserRole update user role
func (us *UserAdminService) UpdateUserRole(ctx context.Context, req *schema.UpdateUserRoleReq) (err error) {
	// Users cannot modify their roles
	if req.UserID == req.LoginUserID {
		return errors.BadRequest(reason.UserCannotUpdateYourRole)
	}

	err = UserRoleRelServicer.SaveUserRole(ctx, req.UserID, req.RoleID)
	if err != nil {
		return err
	}

	//AuthServicer.RemoveUserAllTokens(ctx, req.UserID)
	return
}

// AddUser add user
func (us *UserAdminService) AddUser(ctx context.Context, req *schema.AddUserReq) (err error) {
	_, has, err := repo.UserAdminRepo.GetUserInfoByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if has {
		return errors.BadRequest(reason.EmailDuplicate)
	}

	hashPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userInfo := &entity.User{}
	userInfo.EMail = req.Email
	userInfo.DisplayName = req.DisplayName
	userInfo.Pass = string(hashPwd)

	userInfo.Username, err = UserCommonServicer.MakeUsername(ctx, userInfo.DisplayName)
	if err != nil {
		return err
	}
	userInfo.MailStatus = entity.EmailStatusAvailable
	userInfo.Status = entity.UserStatusAvailable
	userInfo.Rank = 1

	err = repo.UserAdminRepo.AddUser(ctx, userInfo)
	if err != nil {
		return err
	}
	return
}

// AddUsers add users
func (us *UserAdminService) AddUsers(ctx context.Context, req *schema.AddUsersReq) (
	resp []*validator.FormErrorField, err error) {
	resp, err = req.ParseUsers(ctx)
	if err != nil {
		return resp, err
	}
	errData := us.checkUserDuplicateInner(ctx, req.Users)
	if errData != nil {
		return errData.GetErrField(ctx), errors.BadRequest(reason.RequestFormatError)
	}
	users, errData, err := us.formatBulkAddUsers(ctx, req)
	if err != nil {
		return resp, err
	}
	if errData != nil {
		return errData.GetErrField(ctx), errors.BadRequest(reason.RequestFormatError)
	}
	return nil, repo.UserAdminRepo.AddUsers(ctx, users)
}

func (us *UserAdminService) checkUserDuplicateInner(ctx context.Context, users []*schema.AddUserReq) (
	errorData *schema.AddUsersErrorData) {
	lang := utils.GetLangByCtx(ctx)
	val := validator.GetValidatorByLang(lang)

	emails := make(map[string]bool)
	displayNames := make(map[string]bool)
	for line, user := range users {
		if errFields, e := val.Check(user); e != nil {
			errorData = &schema.AddUsersErrorData{}
			if len(errFields) > 0 {
				errorData.Field = errFields[0].ErrorField
				errorData.ExtraMessage = errFields[0].ErrorMsg
			}
			errorData.Line = line + 1
			errorData.Content = fmt.Sprintf("%s, %s, %s", user.DisplayName, user.Email, user.Password)
			return errorData
		}
		if emails[user.Email] {
			return &schema.AddUsersErrorData{
				Field:        "email",
				Line:         line + 1,
				Content:      user.Email,
				ExtraMessage: translator.Tr(lang, reason.EmailDuplicate),
			}
		}
		if displayNames[user.DisplayName] {
			return &schema.AddUsersErrorData{
				Field:        "name",
				Line:         line + 1,
				Content:      user.DisplayName,
				ExtraMessage: translator.Tr(lang, reason.UsernameDuplicate),
			}
		}
		emails[user.Email] = true
		displayNames[user.DisplayName] = true
	}
	return nil
}

func (us *UserAdminService) formatBulkAddUsers(ctx context.Context, req *schema.AddUsersReq) (
	users []*entity.User, errorData *schema.AddUsersErrorData, err error) {
	lang := utils.GetLangByCtx(ctx)
	errorData = &schema.AddUsersErrorData{Line: -1}
	for line, user := range req.Users {
		_, has, e := repo.UserAdminRepo.GetUserInfoByEmail(ctx, user.Email)
		if e != nil {
			return nil, nil, e
		}
		if has {
			errorData.Field = "email"
			errorData.Line = line + 1
			errorData.Content = user.Email
			errorData.ExtraMessage = translator.Tr(lang, reason.EmailDuplicate)
			return nil, errorData, nil
		}

		userInfo := &entity.User{}
		userInfo.EMail = user.Email
		userInfo.DisplayName = user.DisplayName
		hashPwd, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		userInfo.Pass = string(hashPwd)
		userInfo.Username, err = UserCommonServicer.MakeUsername(ctx, userInfo.DisplayName)
		if err != nil {
			errorData.Field = "name"
			errorData.Line = line + 1
			errorData.Content = user.DisplayName
			errorData.ExtraMessage = translator.Tr(lang, reason.UsernameInvalid)
			return nil, errorData, nil
		}
		userInfo.MailStatus = entity.EmailStatusAvailable
		userInfo.Status = entity.UserStatusAvailable
		userInfo.Rank = 1
		users = append(users, userInfo)
	}
	return users, nil, nil
}

// UpdateUserPassword update user password
func (us *UserAdminService) UpdateUserPassword(ctx context.Context, req *schema.UpdateUserPasswordReq) (err error) {
	// Users cannot modify their password
	if req.UserID == req.LoginUserID {
		return errors.BadRequest(reason.AdminCannotUpdateTheirPassword)
	}
	userInfo, exist, err := repo.UserAdminRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}

	hashPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = repo.UserAdminRepo.UpdateUserPassword(ctx, userInfo.ID, string(hashPwd))
	if err != nil {
		return err
	}
	// logout this user
	//AuthServicer.RemoveUserAllTokens(ctx, req.UserID)
	return
}

// GetUserInfo get user one
func (us *UserAdminService) GetUserInfo(ctx context.Context, userID string) (resp *schema.GetUserInfoResp, err error) {
	user, exist, err := repo.UserAdminRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	resp = &schema.GetUserInfoResp{}
	_ = copier.Copy(resp, user)
	return resp, nil
}

// GetUserPage get user list page
func (us *UserAdminService) GetUserPage(ctx context.Context, req *schema.GetUserPageReq) (pageModel *pager.PageModel, err error) {
	user := &entity.User{}
	_ = copier.Copy(user, req)

	if req.IsInactive() {
		user.MailStatus = entity.EmailStatusToBeVerified
		user.Status = entity.UserStatusAvailable
	} else if req.IsSuspended() {
		user.Status = entity.UserStatusSuspended
	} else if req.IsDeleted() {
		user.Status = entity.UserStatusDeleted
	} else {
		user.MailStatus = entity.EmailStatusAvailable
		user.Status = entity.UserStatusAvailable
	}

	if len(req.Query) > 0 {
		if email, e := mail.ParseAddress(req.Query); e == nil {
			user.EMail = email.Address
			req.Query = ""
		} else if strings.HasPrefix(req.Query, "user:") {
			id := strings.TrimSpace(strings.TrimPrefix(req.Query, "user:"))
			idSearch := true
			for _, r := range id {
				if !unicode.IsDigit(r) {
					idSearch = false
					break
				}
			}
			if idSearch {
				user.ID = id
				req.Query = ""
			} else {
				req.Query = id
			}
		}
	}

	users, total, err := repo.UserAdminRepo.GetUserPage(ctx, req.Page, req.PageSize, user, req.Query, req.Staff)
	if err != nil {
		return
	}
	avatarMapping := schema.FormatListAvatar(users)

	resp := make([]*schema.GetUserPageResp, 0)
	for _, u := range users {
		t := &schema.GetUserPageResp{
			UserID:      u.ID,
			CreatedAt:   u.CreatedAt.Unix(),
			Username:    u.Username,
			EMail:       u.EMail,
			Rank:        u.Rank,
			DisplayName: u.DisplayName,
			Avatar:      avatarMapping[u.ID].GetURL(),
		}
		if u.Status == entity.UserStatusDeleted {
			t.Status = constant.UserDeleted
			t.DeletedAt = u.DeletedAt.Unix()
		} else if u.Status == entity.UserStatusSuspended {
			t.Status = constant.UserSuspended
			t.SuspendedAt = u.SuspendedAt.Unix()
		} else if u.MailStatus == entity.EmailStatusToBeVerified {
			t.Status = constant.UserInactive
		} else {
			t.Status = constant.UserNormal
		}
		resp = append(resp, t)
	}
	us.setUserRoleInfo(ctx, resp)
	return pager.NewPageModel(total, resp), nil
}

func (us *UserAdminService) setUserRoleInfo(ctx context.Context, resp []*schema.GetUserPageResp) {
	var userIDs []string
	for _, u := range resp {
		userIDs = append(userIDs, u.UserID)
	}

	userRoleMapping, err := UserRoleRelServicer.GetUserRoleMapping(ctx, userIDs)
	if err != nil {
		glog.Slog.Error(err)
		return
	}

	for _, u := range resp {
		r := userRoleMapping[u.UserID]
		if r == nil {
			continue
		}
		u.RoleID = r.ID
		u.RoleName = r.Name
	}
}

func (us *UserAdminService) GetUserActivation(ctx context.Context, req *schema.GetUserActivationReq) (
	resp *schema.GetUserActivationResp, err error) {
	user, exist, err := repo.UserAdminRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	general := site.Config.GetSiteGeneral()

	data := &schema.EmailCodeContent{
		Email:  user.EMail,
		UserID: user.ID,
	}
	code := uuid.NewString()
	EmailServicer.SaveCode(ctx, code, data.ToJSONString())
	resp = &schema.GetUserActivationResp{
		ActivationURL: fmt.Sprintf("%s/users/account-activation?code=%s", general.SiteUrl, code),
	}
	return resp, nil
}

// SendUserActivation send user activation email
func (us *UserAdminService) SendUserActivation(ctx context.Context, req *schema.SendUserActivationReq) (err error) {
	user, exist, err := repo.UserAdminRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}

	general := site.Config.GetSiteGeneral()
	data := &schema.EmailCodeContent{
		Email:  user.EMail,
		UserID: user.ID,
	}
	code := uuid.NewString()
	EmailServicer.SaveCode(ctx, code, data.ToJSONString())

	verifyEmailURL := fmt.Sprintf("%s/users/account-activation?code=%s", general.SiteUrl, code)
	title, body, err := EmailServicer.RegisterTemplate(ctx, verifyEmailURL)
	if err != nil {
		return err
	}
	go EmailServicer.SendAndSaveCode(ctx, user.EMail, title, body, code, data.ToJSONString())
	return nil
}
