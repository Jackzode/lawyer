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
	notficationcommon "github.com/lawyer/service/notification_common"
	"github.com/segmentfault/pacman/errors"
)

// notificationRepo notification repository
type notificationRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewNotificationRepo new repository
func NewNotificationRepo() notficationcommon.NotificationRepo {
	return &notificationRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddNotification add notification
func (nr *notificationRepo) AddNotification(ctx context.Context, notification *entity.Notification) (err error) {
	notification.ObjectID = uid.DeShortID(notification.ObjectID)
	_, err = nr.DB.Context(ctx).Insert(notification)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) UpdateNotificationContent(ctx context.Context, notification *entity.Notification) (err error) {
	now := time.Now()
	notification.UpdatedAt = now
	notification.ObjectID = uid.DeShortID(notification.ObjectID)
	_, err = nr.DB.Context(ctx).Where("id =?", notification.ID).Cols("content", "updated_at").Update(notification)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) ClearUnRead(ctx context.Context, userID string, notificationType int) (err error) {
	info := &entity.Notification{}
	info.IsRead = schema.NotificationRead
	_, err = nr.DB.Context(ctx).Where("user_id =?", userID).And("type =?", notificationType).Cols("is_read").Update(info)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) ClearIDUnRead(ctx context.Context, userID string, id string) (err error) {
	info := &entity.Notification{}
	info.IsRead = schema.NotificationRead
	_, err = nr.DB.Context(ctx).Where("user_id =?", userID).And("id =?", id).Cols("is_read").Update(info)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) GetById(ctx context.Context, id string) (*entity.Notification, bool, error) {
	info := &entity.Notification{}
	exist, err := nr.DB.Context(ctx).Where("id = ? ", id).Get(info)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return info, false, err
	}
	return info, exist, nil
}

func (nr *notificationRepo) GetByUserIdObjectIdTypeId(ctx context.Context, userID, objectID string, notificationType int) (*entity.Notification, bool, error) {
	info := &entity.Notification{}
	exist, err := nr.DB.Context(ctx).Where("user_id = ? ", userID).And("object_id = ?", objectID).And("type = ?", notificationType).Get(info)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return info, false, err
	}
	return info, exist, nil
}

func (nr *notificationRepo) GetNotificationPage(ctx context.Context, searchCond *schema.NotificationSearch) (
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
