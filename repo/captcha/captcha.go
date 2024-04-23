package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"time"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// captchaRepo captcha repository
type captchaRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewCaptchaRepo new repository
func NewCaptchaRepo() *captchaRepo {
	return &captchaRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (cr *captchaRepo) SetActionType(ctx context.Context, unit, actionType, config string, amount int) (err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", unit, actionType, now.Format("2006-1-02"))
	value := &entity.ActionRecordInfo{}
	value.LastTime = now.Unix()
	value.Num = amount
	valueStr, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	err = cr.Cache.Set(ctx, cacheKey, string(valueStr), 6*time.Minute).Err()
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (cr *captchaRepo) GetActionType(ctx context.Context, Ip, actionType string) (actionInfo *entity.ActionRecordInfo, err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", Ip, actionType, now.Format("2006-1-02"))
	res := cr.Cache.Get(ctx, cacheKey).String()
	if res == "" {
		return nil, nil
	}
	actionInfo = &entity.ActionRecordInfo{}
	_ = json.Unmarshal([]byte(res), actionInfo)
	return actionInfo, nil
}

func (cr *captchaRepo) DelActionType(ctx context.Context, unit, actionType string) (err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", unit, actionType, now.Format("2006-1-02"))
	err = cr.Cache.Del(ctx, cacheKey).Err()
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SetCaptcha set captcha to cache
func (cr *captchaRepo) SetCaptcha(ctx context.Context, key, captcha string) (err error) {
	err = cr.Cache.Set(ctx, key, captcha, 6*time.Minute).Err()
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCaptcha get captcha from cache
func (cr *captchaRepo) GetCaptcha(ctx context.Context, key string) (captcha string, err error) {
	captcha = cr.Cache.Get(ctx, key).String()
	if captcha == "" {
		return "", fmt.Errorf("captcha not exist")
	}
	return captcha, nil
}

func (cr *captchaRepo) DelCaptcha(ctx context.Context, key string) (err error) {
	err = cr.Cache.Del(ctx, key).Err()
	if err != nil {
		log.Debug(err)
	}
	return nil
}
