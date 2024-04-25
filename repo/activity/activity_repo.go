package activity

import (
	"context"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/utils"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// ActivityRepo activity repository
type ActivityRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewActivityRepo new repository
func NewActivityRepo() *ActivityRepo {
	return &ActivityRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (ar *ActivityRepo) GetObjectAllActivity(ctx context.Context, objectID string, showVote bool) (
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

func (ar *ActivityRepo) getAllActivityType(ctx context.Context) (activityTypes []int) {
	var activityTypeNotShown []int
	for _, key := range constant.VoteActivityTypeList {
		id, err := utils.GetIDByKey(ctx, key)
		if err != nil {
			log.Errorf("get config id by key [%s] error: %v", key, err)
		} else {
			activityTypeNotShown = append(activityTypeNotShown, id)
		}
	}
	return activityTypeNotShown
}
