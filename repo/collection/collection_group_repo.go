package collection

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"github.com/segmentfault/pacman/errors"
)

// collectionGroupRepo collectionGroup repository
type collectionGroupRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewCollectionGroupRepo new repository
func NewCollectionGroupRepo() *collectionGroupRepo {
	return &collectionGroupRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddCollectionGroup add collection group
func (cr *collectionGroupRepo) AddCollectionGroup(ctx context.Context, collectionGroup *entity.CollectionGroup) (err error) {
	_, err = cr.DB.Context(ctx).Insert(collectionGroup)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// AddCollectionDefaultGroup add collection group
func (cr *collectionGroupRepo) AddCollectionDefaultGroup(ctx context.Context, userID string) (collectionGroup *entity.CollectionGroup, err error) {
	defaultGroup := &entity.CollectionGroup{
		Name:         "default",
		DefaultGroup: schema.CGDefault,
		UserID:       userID,
	}
	_, err = cr.DB.Context(ctx).Insert(defaultGroup)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	collectionGroup = defaultGroup
	return
}

// CreateDefaultGroupIfNotExist create default group if not exist
func (cr *collectionGroupRepo) CreateDefaultGroupIfNotExist(ctx context.Context, userID string) (
	collectionGroup *entity.CollectionGroup, err error) {
	_, err = cr.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)
		old := &entity.CollectionGroup{
			UserID:       userID,
			DefaultGroup: schema.CGDefault,
		}
		exist, err := session.ForUpdate().Get(old)
		if err != nil {
			return nil, err
		}
		if exist {
			collectionGroup = old
			return old, nil
		}

		defaultGroup := &entity.CollectionGroup{
			Name:         "default",
			DefaultGroup: schema.CGDefault,
			UserID:       userID,
		}
		_, err = session.Insert(defaultGroup)
		if err != nil {
			return nil, err
		}
		collectionGroup = defaultGroup
		return nil, nil
	})
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return collectionGroup, nil
}

// UpdateCollectionGroup update collection group
func (cr *collectionGroupRepo) UpdateCollectionGroup(ctx context.Context, collectionGroup *entity.CollectionGroup, cols []string) (err error) {
	_, err = cr.DB.Context(ctx).ID(collectionGroup.ID).Cols(cols...).Update(collectionGroup)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCollectionGroup get collection group one
func (cr *collectionGroupRepo) GetCollectionGroup(ctx context.Context, id string) (
	collectionGroup *entity.CollectionGroup, exist bool, err error,
) {
	collectionGroup = &entity.CollectionGroup{}
	exist, err = cr.DB.Context(ctx).ID(id).Get(collectionGroup)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCollectionGroupPage get collection group page
func (cr *collectionGroupRepo) GetCollectionGroupPage(ctx context.Context, page, pageSize int, collectionGroup *entity.CollectionGroup) (collectionGroupList []*entity.CollectionGroup, total int64, err error) {
	collectionGroupList = make([]*entity.CollectionGroup, 0)

	session := cr.DB.Context(ctx)
	if collectionGroup.UserID != "" && collectionGroup.UserID != "0" {
		session = session.Where("user_id = ?", collectionGroup.UserID)
	}
	session = session.OrderBy("update_time desc")

	total, err = pager.Help(page, pageSize, collectionGroupList, collectionGroup, session)
	err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	return
}

func (cr *collectionGroupRepo) GetDefaultID(ctx context.Context, userID string) (collectionGroup *entity.CollectionGroup, has bool, err error) {
	collectionGroup = &entity.CollectionGroup{}
	has, err = cr.DB.Context(ctx).Where("user_id =? and  default_group = ?", userID, schema.CGDefault).Get(collectionGroup)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	return
}
