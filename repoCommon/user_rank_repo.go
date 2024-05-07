package repoCommon

import (
	"context"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/utils"
	"github.com/redis/go-redis/v9"

	"github.com/jinzhu/now"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/plugin"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// UserRankRepo user rank repository
type UserRankRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
	//configService *config.ConfigService
}

// NewUserRankRepo new repository
func NewUserRankRepo() *UserRankRepo {
	return &UserRankRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (ur *UserRankRepo) GetMaxDailyRank(ctx context.Context) (maxDailyRank int, err error) {
	maxDailyRank, err = utils.GetIntValue(ctx, "daily_rank_limit")
	if err != nil {
		return 0, err
	}
	return maxDailyRank, nil
}

func (ur *UserRankRepo) CheckReachLimit(ctx context.Context, session *xorm.Session,
	userID string, maxDailyRank int) (
	reach bool, err error) {
	session.Where(builder.Eq{"user_id": userID})
	session.Where(builder.Eq{"cancelled": 0})
	session.Where(builder.Between{
		Col:     "updated_at",
		LessVal: now.BeginningOfDay(),
		MoreVal: now.EndOfDay(),
	})

	earned, err := session.SumInt(&entity.Activity{}, "`rank`")
	if err != nil {
		return false, err
	}
	if int(earned) < maxDailyRank {
		return false, nil
	}
	log.Infof("user %s today has rank %d is reach stand %d", userID, earned, maxDailyRank)
	return true, nil
}

// ChangeUserRank change user rank
func (ur *UserRankRepo) ChangeUserRank(
	ctx context.Context, session *xorm.Session, userID string, userCurrentScore, deltaRank int) (err error) {
	// IMPORTANT: If user center enabled the rank agent, then we should not change user rank.
	//if plugin.RankAgentEnabled() || deltaRank == 0 {
	//	return nil
	//}
	if deltaRank == 0 {
		return nil
	}

	// If user rank is lower than 1 after this action, then user rank will be set to 1 only.
	if deltaRank < 0 && userCurrentScore+deltaRank < 1 {
		deltaRank = 1 - userCurrentScore
	}

	_, err = session.ID(userID).Incr("`rank`", deltaRank).Update(&entity.User{})
	if err != nil {
		return err
	}
	return nil
}

// TriggerUserRank trigger user rank change
// session is need provider, it means this action must be success or failure
// if outer action is failed then this action is need rollback
func (ur *UserRankRepo) TriggerUserRank(ctx context.Context,
	session *xorm.Session, userID string, deltaRank int, activityType int,
) (isReachStandard bool, err error) {
	// IMPORTANT: If user center enabled the rank agent, then we should not change user rank.
	if plugin.RankAgentEnabled() || deltaRank == 0 {
		return false, nil
	}

	if deltaRank < 0 {
		// if user rank is lower than 1 after this action, then user rank will be set to 1 only.
		var isReachMin bool
		isReachMin, err = ur.checkUserMinRank(ctx, session, userID, deltaRank)
		if err != nil {
			return false, err
		}
		if isReachMin {
			_, err = session.Where(builder.Eq{"id": userID}).Update(&entity.User{Rank: 1})
			if err != nil {
				return false, err
			}
			return true, nil
		}
	} else {
		isReachStandard, err = ur.checkUserTodayRank(ctx, session, userID, activityType)
		if err != nil {
			return false, err
		}
		if isReachStandard {
			return isReachStandard, nil
		}
	}
	_, err = session.Where(builder.Eq{"id": userID}).Incr("`rank`", deltaRank).Update(&entity.User{})
	if err != nil {
		return false, err
	}
	return false, nil
}

func (ur *UserRankRepo) checkUserMinRank(ctx context.Context, session *xorm.Session, userID string, deltaRank int) (
	isReachStandard bool, err error,
) {
	bean := &entity.User{ID: userID}
	_, err = session.Select("`rank`").Get(bean)
	if err != nil {
		return false, err
	}
	if bean.Rank+deltaRank < 1 {
		log.Infof("user %s is rank %d out of range before rank operation", userID, deltaRank)
		return true, nil
	}
	return
}

func (ur *UserRankRepo) checkUserTodayRank(ctx context.Context,
	session *xorm.Session, userID string, activityType int,
) (isReachStandard bool, err error) {
	// exclude daily rank
	exclude, _ := utils.GetArrayStringValue(ctx, "daily_rank_limit.exclude")
	for _, item := range exclude {
		cfg, err := utils.GetConfigByKey(ctx, item)
		if err != nil {
			return false, err
		}
		if activityType == cfg.ID {
			return false, nil
		}
	}

	// get user
	start, end := now.BeginningOfDay(), now.EndOfDay()
	session.Where(builder.Eq{"user_id": userID})
	session.Where(builder.Eq{"cancelled": 0})
	session.Where(builder.Between{
		Col:     "updated_at",
		LessVal: start,
		MoreVal: end,
	})
	earned, err := session.Sum(&entity.Activity{}, "`rank`")
	if err != nil {
		return false, err
	}

	// max rank
	maxDailyRank, err := utils.GetIntValue(ctx, "daily_rank_limit")
	if err != nil {
		return false, err
	}

	if int(earned) < maxDailyRank {
		return false, nil
	}
	log.Infof("user %s today has rank %d is reach stand %d", userID, earned, maxDailyRank)
	return true, nil
}

func (ur *UserRankRepo) UserRankPage(ctx context.Context, userID string, page, pageSize int) (
	rankPage []*entity.Activity, total int64, err error,
) {
	rankPage = make([]*entity.Activity, 0)

	session := ur.DB.Context(ctx).Where(builder.Eq{"has_rank": 1}.And(builder.Eq{"cancelled": 0})).And(builder.Gt{"`rank`": 0})
	session.Desc("created_at")

	cond := &entity.Activity{UserID: userID}
	total, err = pager.Help(page, pageSize, &rankPage, cond, session)
	return
}
