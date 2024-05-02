package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"time"
	"xorm.io/xorm"
)

// CaptchaRepo captcha repository
type CaptchaRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewCaptchaRepo new repository
func NewCaptchaRepo() *CaptchaRepo {
	return &CaptchaRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (cr *CaptchaRepo) SetActionType(ctx context.Context, unit, actionType, config string, amount int) (err error) {
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
	return
}

func (cr *CaptchaRepo) GetActionType(ctx context.Context, Ip, actionType string) (actionInfo *entity.ActionRecordInfo, err error) {
	now := time.Now()
	//一天一个key
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", Ip, actionType, now.Format("2006-1-02"))
	res := cr.Cache.Get(ctx, cacheKey).Val()
	if res == "" {
		return nil, nil
	}
	actionInfo = &entity.ActionRecordInfo{}
	_ = json.Unmarshal([]byte(res), actionInfo)
	return actionInfo, nil
}

func (cr *CaptchaRepo) DelActionType(ctx context.Context, unit, actionType string) (err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", unit, actionType, now.Format("2006-1-02"))
	err = cr.Cache.Del(ctx, cacheKey).Err()
	return
}

// SetCaptcha set captcha to cache
func (cr *CaptchaRepo) SetCaptcha(ctx context.Context, key, captcha string) (err error) {
	err = cr.Cache.Set(ctx, key, captcha, 6*time.Minute).Err()
	return
}

// GetCaptcha get captcha from cache
func (cr *CaptchaRepo) GetCaptcha(ctx context.Context, key string) (captcha string, err error) {
	captcha = cr.Cache.Get(ctx, key).Val()
	if captcha == "" {
		return "", fmt.Errorf("captcha not exist")
	}
	return captcha, nil
}

func (cr *CaptchaRepo) DelCaptcha(ctx context.Context, key string) (err error) {
	err = cr.Cache.Del(ctx, key).Err()
	return
}
