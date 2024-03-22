package activity

import (
	"context"
	"fmt"
	"github.com/lawyer/commons/constant/reason"
	entity "github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/utils"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/redis/go-redis/v9"
	"xorm.io/builder"

	"github.com/lawyer/service/activity"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// UserActiveActivityRepo answer accepted
type UserActiveActivityRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

const (
	UserActivated = "user.activated"
)

// NewUserActiveActivityRepo new repository
func NewUserActiveActivityRepo() activity.UserActiveActivityRepo {
	return &UserActiveActivityRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// UserActive user active
func (ar *UserActiveActivityRepo) UserActive(ctx context.Context, userID string) (err error) {
	cfg, err := utils.GetConfigByKey(ctx, UserActivated)
	if err != nil {
		return err
	}
	addActivity := &entity.Activity{
		UserID:           userID,
		ObjectID:         "0",
		OriginalObjectID: "0",
		ActivityType:     cfg.ID,
		Rank:             cfg.GetIntValue(),
		HasRank:          1,
	}

	_, err = ar.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)

		user := &entity.User{}
		exist, err := session.ID(userID).ForUpdate().Get(user)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, fmt.Errorf("user not exist")
		}

		existsActivity := &entity.Activity{}
		exist, err = session.
			And(builder.Eq{"user_id": addActivity.UserID}).
			And(builder.Eq{"activity_type": addActivity.ActivityType}).
			Get(existsActivity)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, nil
		}

		err = repo.UserRankRepo.ChangeUserRank(ctx, session, addActivity.UserID, user.Rank, addActivity.Rank)
		if err != nil {
			return nil, err
		}

		_, err = session.Insert(addActivity)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}
