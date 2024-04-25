package role

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
)

// RolePowerRelRepo rolePowerRel repository
type RolePowerRelRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewRolePowerRelRepo new repository
func NewRolePowerRelRepo() *RolePowerRelRepo {
	return &RolePowerRelRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// GetRolePowerTypeList get role power type list
func (rr *RolePowerRelRepo) GetRolePowerTypeList(ctx context.Context, roleID int) (powers []string, err error) {
	powers = make([]string, 0)
	err = rr.DB.Context(ctx).Table("role_power_rel").
		Cols("power_type").Where(builder.Eq{"role_id": roleID}).Find(&powers)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
