package activity

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	entity2 "github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/lawyer/pkg/obj"
	"github.com/lawyer/service/follow"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"

	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// FollowRepo activity repository
type FollowRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewFollowRepo new repository
func NewFollowRepo() follow.FollowRepo {
	return &FollowRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (ar *FollowRepo) Follow(ctx context.Context, objectID, userID string) error {
	objectTypeStr, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	activityType, err := repo.ActivityRepo.GetActivityTypeByObjectType(ctx, objectTypeStr, "follow")
	if err != nil {
		return err
	}

	_, err = ar.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)
		var (
			existsActivity entity2.Activity
			has            bool
		)
		result = nil

		has, err = session.Where(builder.Eq{"activity_type": activityType}).
			And(builder.Eq{"user_id": userID}).
			And(builder.Eq{"object_id": objectID}).
			Get(&existsActivity)

		if err != nil {
			return
		}

		if has && existsActivity.Cancelled == entity2.ActivityAvailable {
			return
		}

		if has {
			_, err = session.Where(builder.Eq{"id": existsActivity.ID}).
				Cols(`cancelled`).
				Update(&entity2.Activity{
					Cancelled: entity2.ActivityAvailable,
				})
		} else {
			// update existing activity with new user id and u object id
			_, err = session.Insert(&entity2.Activity{
				UserID:           userID,
				ObjectID:         objectID,
				OriginalObjectID: objectID,
				ActivityType:     activityType,
				Cancelled:        entity2.ActivityAvailable,
				Rank:             0,
				HasRank:          0,
			})
		}

		if err != nil {
			log.Error(err)
			return
		}

		// start update followers when everything is fine
		err = ar.updateFollows(ctx, session, objectID, 1)
		if err != nil {
			log.Error(err)
		}

		return
	})

	return err
}

func (ar *FollowRepo) FollowCancel(ctx context.Context, objectID, userID string) error {
	objectTypeStr, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	activityType, err := repo.ActivityRepo.GetActivityTypeByObjectType(ctx, objectTypeStr, "follow")
	if err != nil {
		return err
	}

	_, err = ar.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)
		var (
			existsActivity entity2.Activity
			has            bool
		)
		result = nil

		has, err = session.Where(builder.Eq{"activity_type": activityType}).
			And(builder.Eq{"user_id": userID}).
			And(builder.Eq{"object_id": objectID}).
			Get(&existsActivity)

		if err != nil || !has {
			return
		}

		if has && existsActivity.Cancelled == entity2.ActivityCancelled {
			return
		}
		if _, err = session.Where("id = ?", existsActivity.ID).
			Cols("cancelled").
			Update(&entity2.Activity{
				Cancelled:   entity2.ActivityCancelled,
				CancelledAt: time.Now(),
			}); err != nil {
			return
		}
		err = ar.updateFollows(ctx, session, objectID, -1)
		return
	})
	return err
}

func (ar *FollowRepo) updateFollows(ctx context.Context, session *xorm.Session, objectID string, follows int) error {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return err
	}
	switch objectType {
	case "question":
		_, err = session.Where("id = ?", objectID).Incr("follow_count", follows).Update(&entity2.Question{})
	case "user":
		_, err = session.Where("id = ?", objectID).Incr("follow_count", follows).Update(&entity2.User{})
	case "tag":
		_, err = session.Where("id = ?", objectID).Incr("follow_count", follows).Update(&entity2.Tag{})
	default:
		err = errors.InternalServer(reason.DisallowFollow).WithMsg("this object can't be followed")
	}
	return err
}
