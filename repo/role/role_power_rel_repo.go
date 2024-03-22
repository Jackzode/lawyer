package role

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/lawyer/service/role"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
)

// rolePowerRelRepo rolePowerRel repository
type rolePowerRelRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewRolePowerRelRepo new repository
func NewRolePowerRelRepo() role.RolePowerRelRepo {
	return &rolePowerRelRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// GetRolePowerTypeList get role power type list
func (rr *rolePowerRelRepo) GetRolePowerTypeList(ctx context.Context, roleID int) (powers []string, err error) {
	powers = make([]string, 0)
	err = rr.DB.Context(ctx).Table("role_power_rel").
		Cols("power_type").Where(builder.Eq{"role_id": roleID}).Find(&powers)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
