package limit

import (
	"context"
	"fmt"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/errors"
	"time"
	"xorm.io/xorm"
)

// LimitRepo auth repository
type LimitRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewRateLimitRepo new repository
func NewRateLimitRepo() *LimitRepo {
	return &LimitRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// CheckAndRecord check
func (lr *LimitRepo) CheckAndRecord(ctx context.Context, key string) (limit bool, err error) {
	resp := lr.Cache.Get(ctx, constant.RateLimitCacheKeyPrefix+key).String()
	if resp != "" {
		return true, nil
	}
	err = lr.Cache.Set(ctx, constant.RateLimitCacheKeyPrefix+key,
		fmt.Sprintf("%d", time.Now().Unix()), constant.RateLimitCacheTime).Err()
	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return false, nil
}

// ClearRecord clear
func (lr *LimitRepo) ClearRecord(ctx context.Context, key string) error {
	return lr.Cache.Del(ctx, constant.RateLimitCacheKeyPrefix+key).Err()
}
