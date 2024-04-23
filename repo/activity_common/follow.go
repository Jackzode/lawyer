package activity_common

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	entity2 "github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/repo"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/lawyer/pkg/obj"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// FollowRepo follow repository
type FollowRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewFollowRepo new repository
func NewFollowRepo() *FollowRepo {
	return &FollowRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// GetFollowAmount get object id's follows
func (ar *FollowRepo) GetFollowAmount(ctx context.Context, objectID string) (follows int, err error) {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return 0, err
	}
	switch objectType {
	case "question":
		model := &entity2.Question{}
		_, err = ar.DB.Context(ctx).Where("id = ?", objectID).Cols("`follow_count`").Get(model)
		if err == nil {
			follows = int(model.FollowCount)
		}
	case "user":
		model := &entity2.User{}
		_, err = ar.DB.Context(ctx).Where("id = ?", objectID).Cols("`follow_count`").Get(model)
		if err == nil {
			follows = int(model.FollowCount)
		}
	case "tag":
		model := &entity2.Tag{}
		_, err = ar.DB.Context(ctx).Where("id = ?", objectID).Cols("`follow_count`").Get(model)
		if err == nil {
			follows = int(model.FollowCount)
		}
	default:
		err = errors.InternalServer(reason.DisallowFollow).WithMsg("this object can't be followed")
	}

	if err != nil {
		return 0, err
	}
	return follows, nil
}

// GetFollowUserIDs get follow userID by objectID
func (ar *FollowRepo) GetFollowUserIDs(ctx context.Context, objectID string) (userIDs []string, err error) {
	objectTypeStr, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return nil, err
	}
	activityType, err := repo.ActivityRepo.GetActivityTypeByObjectType(ctx, objectTypeStr, "follow")
	if err != nil {
		log.Errorf("can't get activity type by object key: %s", objectTypeStr)
		return nil, err
	}

	userIDs = make([]string, 0)
	session := ar.DB.Context(ctx).Select("user_id")
	session.Table(entity2.Activity{}.TableName())
	session.Where("object_id = ?", objectID)
	session.Where("activity_type = ?", activityType)
	session.Where("cancelled = 0")
	err = session.Find(&userIDs)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return userIDs, nil
}

// GetFollowIDs get all follow id list
func (ar *FollowRepo) GetFollowIDs(ctx context.Context, userID, objectKey string) (followIDs []string, err error) {
	followIDs = make([]string, 0)
	activityType, err := repo.ActivityRepo.GetActivityTypeByObjectType(ctx, objectKey, "follow")
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	session := ar.DB.Context(ctx).Select("object_id")
	session.Table(entity2.Activity{}.TableName())
	session.Where("user_id = ? AND activity_type = ?", userID, activityType)
	session.Where("cancelled = 0")
	err = session.Find(&followIDs)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return followIDs, nil
}

// IsFollowed check user if follow object or not
func (ar *FollowRepo) IsFollowed(ctx context.Context, userID, objectID string) (followed bool, err error) {
	objectKey, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return false, err
	}

	activityType, err := repo.ActivityRepo.GetActivityTypeByObjectType(ctx, objectKey, "follow")
	if err != nil {
		return false, err
	}

	at := &entity2.Activity{}
	has, err := ar.DB.Context(ctx).Where("user_id = ? AND object_id = ? AND activity_type = ?", userID, objectID, activityType).Get(at)
	if err != nil {
		return false, err
	}
	if !has {
		return false, nil
	}
	if at.Cancelled == entity2.ActivityCancelled {
		return false, nil
	} else {
		return true, nil
	}
}
