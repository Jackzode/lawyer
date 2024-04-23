package activity_common

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/pkg/uid"
	"github.com/lawyer/repo"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// VoteRepo activity repository
type VoteRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewVoteRepo new repository
func NewVoteRepo() *VoteRepo {
	return &VoteRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (vr *VoteRepo) GetVoteStatus(ctx context.Context, objectID, userID string) (status string) {

	objectID = uid.DeShortID(objectID)
	for _, action := range []string{"vote_up", "vote_down"} {
		activityType, _, _, err := repo.ActivityRepo.GetActivityTypeByObjID(ctx, objectID, action)
		if err != nil {
			return ""
		}
		at := &entity.Activity{}
		has, err := vr.DB.Context(ctx).Where("object_id = ? AND cancelled = 0 AND activity_type = ? AND user_id = ?",
			objectID, activityType, userID).Get(at)
		if err != nil {
			log.Error(err)
			return ""
		}
		if has {
			return action
		}
	}
	return ""
}

func (vr *VoteRepo) GetVoteCount(ctx context.Context, activityTypes []int) (count int64, err error) {
	list := make([]*entity.Activity, 0)
	count, err = vr.DB.Context(ctx).Where("cancelled =0").In("activity_type", activityTypes).FindAndCount(&list)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
