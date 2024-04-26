package repoCommon

import (
	"context"
	"fmt"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/utils"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/lawyer/pkg/obj"
	"xorm.io/builder"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
)

// ActivityComRepo activity repository
type ActivityComRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewActivityRepo new repository
func NewActivityRepo() *ActivityComRepo {
	return &ActivityComRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (ar *ActivityComRepo) GetActivityTypeByObjID(ctx context.Context, objectID string, action string) (
	activityType, rank, hasRank int, err error) {

	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return
	}

	confKey := fmt.Sprintf("%s.%s", objectType, action)
	cfg, err := utils.GetConfigByKey(ctx, confKey)
	if err != nil {
		return
	}
	activityType, rank = cfg.ID, cfg.GetIntValue()
	hasRank = 0
	if rank != 0 {
		hasRank = 1
	}
	return
}

func (ar *ActivityComRepo) GetActivityTypeByObjectType(ctx context.Context, objectType, action string) (activityType int, err error) {
	configKey := fmt.Sprintf("%s.%s", objectType, action)
	cfg, err := utils.GetConfigByKey(ctx, configKey)
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return cfg.ID, nil
}

func (ar *ActivityComRepo) GetActivityTypeByConfigKey(ctx context.Context, configKey string) (activityType int, err error) {
	cfg, err := utils.GetConfigByKey(ctx, configKey)
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return cfg.ID, nil
}

func (ar *ActivityComRepo) GetActivity(ctx context.Context, session *xorm.Session,
	objectID, userID string, activityType int,
) (existsActivity *entity.Activity, exist bool, err error) {
	existsActivity = &entity.Activity{}
	exist, err = session.
		Where(builder.Eq{"object_id": objectID}).
		And(builder.Eq{"user_id": userID}).
		And(builder.Eq{"activity_type": activityType}).
		Get(existsActivity)
	return
}

func (ar *ActivityComRepo) GetUserIDObjectIDActivitySum(ctx context.Context, userID, objectID string) (int, error) {
	sum := &entity.ActivityRankSum{}
	_, err := ar.DB.Context(ctx).Table(entity.Activity{}.TableName()).
		Select("sum(`rank`) as `rank`").
		Where("user_id =?", userID).
		And("object_id = ?", objectID).
		And("cancelled =0").
		Get(sum)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return 0, err
	}
	return sum.Rank, nil
}

// AddActivity add activity
func (ar *ActivityComRepo) AddActivity(ctx context.Context, activity *entity.Activity) (err error) {
	_, err = ar.DB.Context(ctx).Insert(activity)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUsersWhoHasGainedTheMostReputation get users who has gained the most reputation over a period of time
func (ar *ActivityComRepo) GetUsersWhoHasGainedTheMostReputation(
	ctx context.Context, startTime, endTime time.Time, limit int) (rankStat []*entity.ActivityUserRankStat, err error) {
	rankStat = make([]*entity.ActivityUserRankStat, 0)
	session := ar.DB.Context(ctx).Select("user_id, SUM(`rank`) AS rank_amount").Table("activity")
	session.Where("has_rank = 1 AND cancelled = 0")
	session.Where("created_at >= ?", startTime)
	session.Where("created_at <= ?", endTime)
	session.GroupBy("user_id")
	session.Desc("rank_amount")
	session.Limit(limit)
	err = session.Find(&rankStat)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUsersWhoHasVoteMost get users who has vote most
func (ar *ActivityComRepo) GetUsersWhoHasVoteMost(
	ctx context.Context, startTime, endTime time.Time, limit int) (voteStat []*entity.ActivityUserVoteStat, err error) {
	voteStat = make([]*entity.ActivityUserVoteStat, 0)

	actIDs := make([]int, 0)
	for _, act := range constant.ActivityTypeList {
		cfg, err := utils.GetConfigByKey(ctx, act)
		if err == nil {
			actIDs = append(actIDs, cfg.ID)
		}
	}

	session := ar.DB.Context(ctx).Select("user_id, COUNT(*) AS vote_count").Table("activity")
	session.Where("cancelled = 0")
	session.In("activity_type", actIDs)
	session.Where("created_at >= ?", startTime)
	session.Where("created_at <= ?", endTime)
	session.GroupBy("user_id")
	session.Desc("vote_count")
	session.Limit(limit)
	err = session.Find(&voteStat)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
