package notification

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"time"
	"xorm.io/xorm"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
)

// NotificationRepo notification repository
type NotificationRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewNotificationRepo new repository
func NewNotificationRepo() *NotificationRepo {
	return &NotificationRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddNotification add notification
func (nr *NotificationRepo) AddNotification(ctx context.Context, notification *entity.Notification) (err error) {
	notification.ObjectID = uid.DeShortID(notification.ObjectID)
	_, err = nr.DB.Context(ctx).Insert(notification)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *NotificationRepo) UpdateNotificationContent(ctx context.Context, notification *entity.Notification) (err error) {
	now := time.Now()
	notification.UpdatedAt = now
	notification.ObjectID = uid.DeShortID(notification.ObjectID)
	_, err = nr.DB.Context(ctx).Where("id =?", notification.ID).Cols("content", "updated_at").Update(notification)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *NotificationRepo) ClearUnRead(ctx context.Context, userID string, notificationType int) (err error) {
	info := &entity.Notification{}
	info.IsRead = schema.NotificationRead
	_, err = nr.DB.Context(ctx).Where("user_id =?", userID).And("type =?", notificationType).Cols("is_read").Update(info)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *NotificationRepo) ClearIDUnRead(ctx context.Context, userID string, id string) (err error) {
	info := &entity.Notification{}
	info.IsRead = schema.NotificationRead
	_, err = nr.DB.Context(ctx).Where("user_id =?", userID).And("id =?", id).Cols("is_read").Update(info)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *NotificationRepo) GetById(ctx context.Context, id string) (*entity.Notification, bool, error) {
	info := &entity.Notification{}
	exist, err := nr.DB.Context(ctx).Where("id = ? ", id).Get(info)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return info, false, err
	}
	return info, exist, nil
}

func (nr *NotificationRepo) GetByUserIdObjectIdTypeId(ctx context.Context, userID, objectID string, notificationType int) (*entity.Notification, bool, error) {
	info := &entity.Notification{}
	exist, err := nr.DB.Context(ctx).Where("user_id = ? ", userID).And("object_id = ?", objectID).And("type = ?", notificationType).Get(info)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return info, false, err
	}
	return info, exist, nil
}

func (nr *NotificationRepo) GetNotificationPage(ctx context.Context, searchCond *schema.NotificationSearch) (
	notificationList []*entity.Notification, total int64, err error) {
	notificationList = make([]*entity.Notification, 0)
	if searchCond.UserID == "" {
		return notificationList, 0, nil
	}

	session := nr.DB.Context(ctx)
	session = session.Desc("updated_at")

	cond := &entity.Notification{
		UserID: searchCond.UserID,
		Type:   searchCond.Type,
	}
	if searchCond.InboxType > 0 {
		cond.MsgType = searchCond.InboxType
	}
	total, err = pager.Help(searchCond.Page, searchCond.PageSize, &notificationList, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
