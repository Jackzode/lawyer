package unique

import (
	"context"
	"fmt"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/lawyer/service/unique"
	"github.com/segmentfault/pacman/errors"
)

// uniqueIDRepo Unique id repository
type uniqueIDRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewUniqueIDRepo new repository
func NewUniqueIDRepo() unique.UniqueIDRepo {
	return &uniqueIDRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// GenUniqueIDStr generate unique id string
// 1 + 00x(objectType) + 000000000000x(id)
func (ur *uniqueIDRepo) GenUniqueIDStr(ctx context.Context, key string) (uniqueID string, err error) {
	objectType := constant.ObjectTypeStrMapping[key]
	bean := &entity.Uniqid{UniqidType: objectType}
	_, err = ur.DB.Context(ctx).Insert(bean)
	if err != nil {
		return "", errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return fmt.Sprintf("1%03d%013d", objectType, bean.ID), nil
}
