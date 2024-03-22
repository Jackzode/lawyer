package role

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/lawyer/service/role"
	"github.com/segmentfault/pacman/errors"
)

// powerRepo power repository
type powerRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewPowerRepo new repository
func NewPowerRepo(DB *xorm.Engine, Cache *redis.Client) role.PowerRepo {
	return &powerRepo{
		DB:    DB,
		Cache: Cache,
	}
}

// GetPowerList get  list all
func (pr *powerRepo) GetPowerList(ctx context.Context, power *entity.Power) (powerList []*entity.Power, err error) {
	powerList = make([]*entity.Power, 0)
	err = pr.DB.Context(ctx).Find(powerList, power)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
