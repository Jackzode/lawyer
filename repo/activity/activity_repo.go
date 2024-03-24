package activity

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/utils"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/lawyer/service/activity"
	"github.com/lawyer/service/activity_type"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// activityRepo activity repository
type activityRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewActivityRepo new repository
func NewActivityRepo() activity.ActivityRepo {
	return &activityRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (ar *activityRepo) GetObjectAllActivity(ctx context.Context, objectID string, showVote bool) (
	activityList []*entity.Activity, err error) {
	activityList = make([]*entity.Activity, 0)
	session := ar.DB.Context(ctx).Desc("id")

	if !showVote {
		activityTypeNotShown := ar.getAllActivityType(ctx)
		session.NotIn("activity_type", activityTypeNotShown)
	}
	err = session.Find(&activityList, &entity.Activity{OriginalObjectID: objectID})
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return activityList, nil
}

func (ar *activityRepo) getAllActivityType(ctx context.Context) (activityTypes []int) {
	var activityTypeNotShown []int
	for _, key := range activity_type.VoteActivityTypeList {
		id, err := utils.GetIDByKey(ctx, key)
		if err != nil {
			log.Errorf("get config id by key [%s] error: %v", key, err)
		} else {
			activityTypeNotShown = append(activityTypeNotShown, id)
		}
	}
	return activityTypeNotShown
}