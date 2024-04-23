package user_notification_config

import (
	"context"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// userNotificationConfigRepo notification repository
type userNotificationConfigRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewUserNotificationConfigRepo new repository
func NewUserNotificationConfigRepo() *userNotificationConfigRepo {
	return &userNotificationConfigRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// Add add notification config
func (ur *userNotificationConfigRepo) Add(ctx context.Context, userIDs []string, source, channels string) (err error) {
	var configs []*entity.UserNotificationConfig
	for _, userID := range userIDs {
		configs = append(configs, &entity.UserNotificationConfig{
			UserID:   userID,
			Source:   source,
			Channels: channels,
			Enabled:  true,
		})
	}
	_, err = ur.DB.Context(ctx).Insert(configs)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// Save save notification config, if existed, update, if not exist, insert
func (ur *userNotificationConfigRepo) Save(ctx context.Context, uc *entity.UserNotificationConfig) (err error) {
	old := &entity.UserNotificationConfig{UserID: uc.UserID, Source: uc.Source}
	exist, err := ur.DB.Context(ctx).Get(old)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		old.Channels = uc.Channels
		old.Enabled = uc.Enabled
		_, err = ur.DB.Context(ctx).ID(old.ID).UseBool("enabled").Cols("channels", "enabled").Update(old)
	} else {
		_, err = ur.DB.Context(ctx).Insert(uc)
	}
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// GetByUserID get notification config by user id
func (ur *userNotificationConfigRepo) GetByUserID(ctx context.Context, userID string) (
	[]*entity.UserNotificationConfig, error) {
	var configs []*entity.UserNotificationConfig
	err := ur.DB.Context(ctx).Where("user_id = ?", userID).Find(&configs)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return configs, nil
}

// GetBySource get notification config by source
func (ur *userNotificationConfigRepo) GetBySource(ctx context.Context, source constant.NotificationSource) (
	[]*entity.UserNotificationConfig, error) {
	var configs []*entity.UserNotificationConfig
	err := ur.DB.Context(ctx).UseBool("enabled").
		Find(&configs, &entity.UserNotificationConfig{Source: string(source), Enabled: true})
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return configs, nil
}

// GetByUserIDAndSource get notification config by user id and source
func (ur *userNotificationConfigRepo) GetByUserIDAndSource(ctx context.Context, userID string, source constant.NotificationSource) (
	conf *entity.UserNotificationConfig, exist bool, err error) {
	config := &entity.UserNotificationConfig{UserID: userID, Source: string(source)}
	exist, err = ur.DB.Context(ctx).Get(config)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return config, exist, nil
}

// GetByUsersAndSource get notification config by user ids and source
func (ur *userNotificationConfigRepo) GetByUsersAndSource(
	ctx context.Context, userIDs []string, source constant.NotificationSource) (
	[]*entity.UserNotificationConfig, error) {
	var configs []*entity.UserNotificationConfig
	err := ur.DB.Context(ctx).UseBool("enabled").In("user_id", userIDs).
		Find(&configs, &entity.UserNotificationConfig{Source: string(source), Enabled: true})
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return configs, nil
}
