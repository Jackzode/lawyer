package user

import (
	"context"
	"encoding/json"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/repo/auth"
	"github.com/redis/go-redis/v9"
	"time"
	"xorm.io/xorm"

	"xorm.io/builder"

	"github.com/lawyer/commons/utils/pager"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// userAdminRepo user repository
type userAdminRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewUserAdminRepo new repository
func NewUserAdminRepo() *userAdminRepo {
	return &userAdminRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// UpdateUserStatus update user status
func (ur *userAdminRepo) UpdateUserStatus(ctx context.Context, userID string, userStatus, mailStatus int,
	email string,
) (err error) {
	cond := &entity.User{Status: userStatus, MailStatus: mailStatus, EMail: email}
	switch userStatus {
	case entity.UserStatusSuspended:
		cond.SuspendedAt = time.Now()
	case entity.UserStatusDeleted:
		cond.DeletedAt = time.Now()
	}
	_, err = ur.DB.Context(ctx).ID(userID).Update(cond)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	userCacheInfo := &entity.UserCacheInfo{
		UserID:      userID,
		EmailStatus: mailStatus,
		UserStatus:  userStatus,
	}
	t, _ := json.Marshal(userCacheInfo)
	log.Infof("user change status: %s", string(t))
	err = auth.NewAuthRepo().SetUserStatus(ctx, userID, userCacheInfo)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// AddUser add user
func (ur *userAdminRepo) AddUser(ctx context.Context, user *entity.User) (err error) {
	_, err = ur.DB.Context(ctx).Insert(user)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// AddUsers add users
func (ur *userAdminRepo) AddUsers(ctx context.Context, users []*entity.User) (err error) {
	_, err = ur.DB.Context(ctx).Insert(users)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateUserPassword update user password
func (ur *userAdminRepo) UpdateUserPassword(ctx context.Context, userID string, password string) (err error) {
	_, err = ur.DB.Context(ctx).ID(userID).Update(&entity.User{Pass: password})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUserInfo get user info
func (ur *userAdminRepo) GetUserInfo(ctx context.Context, userID string) (user *entity.User, exist bool, err error) {
	user = &entity.User{}
	exist, err = ur.DB.Context(ctx).ID(userID).Get(user)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return
	}
	err = tryToDecorateUserInfoFromUserCenter(ctx, ur.DB, user)
	if err != nil {
		return nil, false, err
	}
	return
}

// GetUserInfoByEmail get user info
func (ur *userAdminRepo) GetUserInfoByEmail(ctx context.Context, email string) (user *entity.User, exist bool, err error) {
	userInfo := &entity.User{}
	exist, err = ur.DB.Context(ctx).Where("e_mail = ?", email).
		Where("status != ?", entity.UserStatusDeleted).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if !exist {
		return
	}
	err = tryToDecorateUserInfoFromUserCenter(ctx, ur.DB, user)
	if err != nil {
		return nil, false, err
	}
	return
}

// GetUserPage get user page
func (ur *userAdminRepo) GetUserPage(ctx context.Context, page, pageSize int, user *entity.User,
	usernameOrDisplayName string, isStaff bool) (users []*entity.User, total int64, err error) {
	users = make([]*entity.User, 0)
	session := ur.DB.Context(ctx)
	switch user.Status {
	case entity.UserStatusDeleted:
		session.Desc("`user`.deleted_at")
	case entity.UserStatusSuspended:
		session.Desc("`user`.suspended_at")
	default:
		session.Desc("`user`.created_at")
	}

	if len(usernameOrDisplayName) > 0 {
		session.And(builder.Or(
			builder.Like{"`user`.username", usernameOrDisplayName},
			builder.Like{"`user`.display_name", usernameOrDisplayName},
		))
	}
	if isStaff {
		session.Join("INNER", "user_role_rel", "`user`.id = `user_role_rel`.user_id AND `user_role_rel`.role_id > 1")
	}

	total, err = pager.Help(page, pageSize, &users, user, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	tryToDecorateUserListFromUserCenter(ctx, ur.DB, users)
	return
}
