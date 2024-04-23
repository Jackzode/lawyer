package export

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"time"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
)

// emailRepo email repository
type emailRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewEmailRepo new repository
func NewEmailRepo() *emailRepo {
	return &emailRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// SetCode The email code is used to verify that the link in the message is out of date
func (e *emailRepo) SetCode(ctx context.Context, code, content string, duration time.Duration) error {
	err := e.Cache.Set(ctx, code, content, duration).Err()
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// VerifyCode verify the code if out of date
func (e *emailRepo) VerifyCode(ctx context.Context, code string) (content string, err error) {
	content = e.Cache.Get(ctx, code).String()
	return content, nil
}
