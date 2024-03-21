package unique

import (
	"context"
	"fmt"
	"github.com/apache/incubator-answer/commons/constant"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/commons/entity"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/apache/incubator-answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
)

// uniqueIDRepo Unique id repository
type uniqueIDRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewUniqueIDRepo new repository
func NewUniqueIDRepo(DB *xorm.Engine, Cache *redis.Client) unique.UniqueIDRepo {
	return &uniqueIDRepo{
		DB:    DB,
		Cache: Cache,
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
