package user

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	glog "github.com/lawyer/commons/logger"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/plugin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

// UserRepo user repository
type UserRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewUserRepo new repository
func NewUserRepo() *UserRepo {
	return &UserRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddUser add user
func (ur *UserRepo) AddUser(ctx context.Context, user *entity.User) (err error) {

	_, err = ur.DB.Transaction(
		func(session *xorm.Session) (interface{}, error) {
			session = session.Context(ctx)
			userInfo := &entity.User{}
			exist, err := session.Where("username = ?", user.Username).Get(userInfo)
			if err != nil {
				return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
			}
			if exist {
				return nil, errors.InternalServer(reason.UsernameDuplicate)
			}
			_, err = session.Insert(user)
			if err != nil {
				return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
			}
			return nil, nil
		})
	return
}

// IncreaseAnswerCount increase answer count
func (ur *UserRepo) IncreaseAnswerCount(ctx context.Context, userID string, amount int) (err error) {
	user := &entity.User{}
	_, err = ur.DB.Context(ctx).Where("id = ?", userID).Incr("answer_count", amount).Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// IncreaseQuestionCount increase question count
func (ur *UserRepo) IncreaseQuestionCount(ctx context.Context, userID string, amount int) (err error) {
	user := &entity.User{}
	_, err = ur.DB.Context(ctx).Where("id = ?", userID).Incr("question_count", amount).Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *UserRepo) UpdateQuestionCount(ctx context.Context, userID string, count int64) (err error) {
	user := &entity.User{}
	user.QuestionCount = int(count)
	_, err = ur.DB.Context(ctx).Where("id = ?", userID).Cols("question_count").Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *UserRepo) UpdateAnswerCount(ctx context.Context, userID string, count int) (err error) {
	user := &entity.User{}
	user.AnswerCount = count
	_, err = ur.DB.Context(ctx).Where("id = ?", userID).Cols("answer_count").Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// UpdateLastLoginDate update last login date
func (ur *UserRepo) UpdateLastLoginDate(ctx context.Context, userID string) (err error) {
	user := &entity.User{LastLoginDate: time.Now()}
	_, err = ur.DB.Context(ctx).Where("id = ?", userID).Cols("last_login_date").Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// UpdateEmailStatus update email status
func (ur *UserRepo) UpdateEmailStatus(ctx context.Context, userID string, emailStatus int) error {
	cond := &entity.User{MailStatus: emailStatus}
	_, err := ur.DB.Context(ctx).Where("id = ?", userID).Cols("mail_status").Update(cond)
	if err != nil {
		return err
	}
	return nil
}

// UpdateNoticeStatus update notice status
func (ur *UserRepo) UpdateNoticeStatus(ctx context.Context, userID string, noticeStatus int) error {
	cond := &entity.User{NoticeStatus: noticeStatus}
	_, err := ur.DB.Context(ctx).Where("id = ?", userID).Cols("notice_status").Update(cond)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *UserRepo) UpdatePass(ctx context.Context, userID, pass string) error {
	_, err := ur.DB.Context(ctx).Where("id = ?", userID).Cols("pass").Update(&entity.User{Pass: pass})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *UserRepo) UpdateEmail(ctx context.Context, userID, email string) (err error) {
	_, err = ur.DB.Context(ctx).Where("id = ?", userID).Update(&entity.User{EMail: email})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ur *UserRepo) UpdateEmailAndEmailStatus(ctx context.Context, userID, email string, mailStatus int) (err error) {
	_, err = ur.DB.Context(ctx).Where("id = ?", userID).Update(&entity.User{EMail: email, MailStatus: mailStatus})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ur *UserRepo) UpdateLanguage(ctx context.Context, userID, language string) (err error) {
	_, err = ur.DB.Context(ctx).Where("id = ?", userID).Update(&entity.User{Language: language})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateInfo update user info
func (ur *UserRepo) UpdateInfo(ctx context.Context, userInfo *entity.User) (err error) {
	_, err = ur.DB.Context(ctx).Where("id = ?", userInfo.ID).
		Cols("username", "display_name", "avatar", "bio", "bio_html", "website", "location").Update(userInfo)
	return
}

// GetByUserID get user info by user id
func (ur *UserRepo) GetByUserID(ctx context.Context, userID string) (userInfo *entity.User, exist bool, err error) {
	userInfo = &entity.User{}
	exist, err = ur.DB.Context(ctx).Where("id = ?", userID).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	//todo  due to plugin
	err = tryToDecorateUserInfoFromUserCenter(ctx, ur.DB, userInfo)
	if err != nil {
		return nil, false, err
	}
	return
}

func (ur *UserRepo) BatchGetByID(ctx context.Context, ids []string) ([]*entity.User, error) {
	list := make([]*entity.User, 0)
	err := ur.DB.Context(ctx).In("id", ids).Find(&list)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	tryToDecorateUserListFromUserCenter(ctx, ur.DB, list)
	return list, nil
}

// GetByUsername get user by username
func (ur *UserRepo) GetUserInfoByUsername(ctx context.Context, username string) (userInfo *entity.User, exist bool, err error) {
	userInfo = &entity.User{}
	exist, err = ur.DB.Context(ctx).Where("username = ?", username).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	//todo
	//err = tryToDecorateUserInfoFromUserCenter(ctx, ur.DB, userInfo)
	//if err != nil {
	//	return nil, false, err
	//}
	return
}

func (ur *UserRepo) GetByUsernames(ctx context.Context, usernames []string) ([]*entity.User, error) {
	list := make([]*entity.User, 0)
	err := ur.DB.Context(ctx).Where("status =?", entity.UserStatusAvailable).In("username", usernames).Find(&list)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return list, err
	}
	tryToDecorateUserListFromUserCenter(ctx, ur.DB, list)
	return list, nil
}

// GetByEmail get user by email
func (ur *UserRepo) GetUserInfoByEmailFromDB(ctx context.Context, email string) (userInfo *entity.User, exist bool, err error) {
	userInfo = &entity.User{}
	exist, err = ur.DB.Context(ctx).Where("e_mail = ?", email).
		Where("status != ?", entity.UserStatusDeleted).Get(userInfo)
	if err != nil {
		glog.Slog.Error(err.Error())
		//err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ur *UserRepo) GetUserCount(ctx context.Context) (count int64, err error) {
	session := ur.DB.Context(ctx)
	session.Where("status = ? OR status = ?", entity.UserStatusAvailable, entity.UserStatusSuspended)
	count, err = session.Count(&entity.User{})
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (ur *UserRepo) SearchUserListByName(ctx context.Context, name string, limit int) (userList []*entity.User, err error) {
	userList = make([]*entity.User, 0)
	session := ur.DB.Context(ctx)
	session.Where("status = ?", entity.UserStatusAvailable)
	session.Where("username LIKE ? OR display_name LIKE ?", strings.ToLower(name)+"%", name+"%")
	session.OrderBy("username ASC, id DESC")
	session.Limit(limit)
	err = session.Find(&userList)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	//todo
	//tryToDecorateUserListFromUserCenter(ctx, ur.DB, userList)
	return
}

func tryToDecorateUserInfoFromUserCenter(ctx context.Context, db *xorm.Engine, original *entity.User) (err error) {
	if original == nil {
		return nil
	}
	uc, ok := plugin.GetUserCenter()
	if !ok {
		return nil
	}

	userInfo := &entity.UserExternalLogin{}
	session := db.Context(ctx).Where("user_id = ?", original.ID)
	session.Where("provider = ?", uc.Info().SlugName)
	exist, err := session.Get(userInfo)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil
	}

	userCenterBasicUserInfo, err := uc.UserInfo(userInfo.ExternalID)
	if err != nil {
		log.Error(err)
		return errors.BadRequest(reason.UserNotFound).WithError(err).WithStack()
	}

	// In general, usernames should be guaranteed unique by the User Center plugin, so there are no inconsistencies.
	if original.Username != userCenterBasicUserInfo.Username {
		log.Warnf("user %s username is inconsistent with user center", original.ID)
	}
	decorateByUserCenterUser(original, userCenterBasicUserInfo)
	return nil
}

func tryToDecorateUserListFromUserCenter(ctx context.Context, db *xorm.Engine, original []*entity.User) {
	uc, ok := plugin.GetUserCenter()
	if !ok {
		return
	}

	ids := make([]string, 0)
	originalUserIDMapping := make(map[string]*entity.User, 0)
	for _, user := range original {
		originalUserIDMapping[user.ID] = user
		ids = append(ids, user.ID)
	}

	userExternalLoginList := make([]*entity.UserExternalLogin, 0)
	session := db.Context(ctx).Where("provider = ?", uc.Info().SlugName)
	session.In("user_id", ids)
	err := session.Find(&userExternalLoginList)
	if err != nil {
		log.Error(err)
		return
	}

	userExternalIDs := make([]string, 0)
	originalExternalIDMapping := make(map[string]*entity.User, 0)
	for _, u := range userExternalLoginList {
		originalExternalIDMapping[u.ExternalID] = originalUserIDMapping[u.UserID]
		userExternalIDs = append(userExternalIDs, u.ExternalID)
	}
	if len(userExternalIDs) == 0 {
		return
	}

	ucUsers, err := uc.UserList(userExternalIDs)
	if err != nil {
		log.Errorf("get user list from user center failed: %v, %v", err, userExternalIDs)
		return
	}

	for _, ucUser := range ucUsers {
		decorateByUserCenterUser(originalExternalIDMapping[ucUser.ExternalID], ucUser)
	}
}

func decorateByUserCenterUser(original *entity.User, ucUser *plugin.UserCenterBasicUserInfo) {
	if original == nil || ucUser == nil {
		return
	}
	// In general, usernames should be guaranteed unique by the User Center plugin, so there are no inconsistencies.
	if original.Username != ucUser.Username {
		log.Warnf("user %s username is inconsistent with user center", original.ID)
	}
	if len(ucUser.DisplayName) > 0 {
		original.DisplayName = ucUser.DisplayName
	}
	if len(ucUser.Email) > 0 {
		original.EMail = ucUser.Email
	}
	if len(ucUser.Avatar) > 0 {
		original.Avatar = schema.CustomAvatar(ucUser.Avatar).ToJsonString()
	}
	if len(ucUser.Mobile) > 0 {
		original.Mobile = ucUser.Mobile
	}
	if len(ucUser.Bio) > 0 {
		original.BioHTML = converter.Markdown2HTML(ucUser.Bio) + original.BioHTML
	}

	// If plugin enable rank agent, use rank from user center.
	if plugin.RankAgentEnabled() {
		original.Rank = ucUser.Rank
	}
	if ucUser.Status != plugin.UserStatusAvailable {
		original.Status = int(ucUser.Status)
	}
}
