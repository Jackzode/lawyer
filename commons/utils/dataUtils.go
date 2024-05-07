package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	glog "github.com/lawyer/commons/logger"
)

func GetConfigByID(ctx context.Context, id int) (c *entity.Config, err error) {
	cacheKey := fmt.Sprintf("%s%d", constant.ConfigID2KEYCacheKeyPrefix, id)
	cachehandler := handler.RedisClient.Get(ctx, cacheKey).String()
	if len(cachehandler) > 0 {
		c = &entity.Config{}
		c.BuildByJSON([]byte(cachehandler))
		if c.ID > 0 {
			return c, nil
		}
	}

	c = &entity.Config{}
	exist, err := handler.Engine.Context(ctx).ID(id).Get(c)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("config not found by id: %d", id)
	}

	// update cache
	if err := handler.RedisClient.Set(ctx, cacheKey, c.JsonString(), constant.ConfigCacheTime).Err(); err != nil {
		glog.Slog.Error(err)
	}
	return c, nil
}

func GetConfigByKey(ctx context.Context, key string) (c *entity.Config, err error) {
	cacheKey := constant.ConfigKEY2ContentCacheKeyPrefix + key
	cachehandler := handler.RedisClient.Get(ctx, cacheKey).Val()
	if len(cachehandler) > 0 {
		c = &entity.Config{}
		c.BuildByJSON([]byte(cachehandler))
		if c.ID > 0 {
			return c, nil
		}
	}

	c = &entity.Config{Key: key}
	exist, err := handler.Engine.Context(ctx).Get(c)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("config not found by key: %s", key)
	}

	// update cache
	err = handler.RedisClient.Set(ctx, cacheKey, c.JsonString(), constant.ConfigCacheTime).Err()
	if err != nil {
		glog.Slog.Error(err)
	}
	return c, nil
}

func UpdateConfig(ctx context.Context, key string, value string) (err error) {
	// check if key exists
	oldConfig := &entity.Config{Key: key}
	exist, err := handler.Engine.Context(ctx).Get(oldConfig)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New(reason.ObjectNotFound)
	}

	// update handlerbase
	_, err = handler.Engine.Context(ctx).ID(oldConfig.ID).Update(&entity.Config{Value: value})
	if err != nil {
		return err
	}

	//delete cache
	err = handler.RedisClient.Del(ctx, constant.ConfigKEY2ContentCacheKeyPrefix+key).Err()
	if err != nil {
		glog.Slog.Error(err)
	}
	err = handler.RedisClient.Del(ctx, fmt.Sprintf("%s%d", constant.ConfigID2KEYCacheKeyPrefix, oldConfig.ID)).Err()
	if err != nil {
		glog.Slog.Error(err)
	}
	return
	// update cache 改为删除缓存
	//oldConfig.Value = value
	//cacheVal := oldConfig.JsonString()
	//if err = handler.RedisClient.Set(ctx,
	//	constant.ConfigKEY2ContentCacheKeyPrefix+key, cacheVal, constant.ConfigCacheTime).Err(); err != nil {
	//	glog.Slog.Error(err)
	//}
	//if err = handler.RedisClient.Set(ctx,
	//	fmt.Sprintf("%s%d", constant.ConfigID2KEYCacheKeyPrefix, oldConfig.ID), cacheVal, constant.ConfigCacheTime).Err(); err != nil {
	//	glog.Slog.Error(err)
	//}

}

// GetIntValue get config int value
func GetIntValue(ctx context.Context, key string) (val int, err error) {
	cf, err := GetConfigByKey(ctx, key)
	if err != nil {
		return 0, err
	}
	return cf.GetIntValue(), nil
}

// GetStringValue get config string value
func GetStringValue(ctx context.Context, key string) (val string, err error) {
	cf, err := GetConfigByKey(ctx, key)
	if err != nil {
		return "", err
	}
	return cf.Value, nil
}

// GetArrayStringValue get config array string value
func GetArrayStringValue(ctx context.Context, key string) (val []string, err error) {
	cf, err := GetConfigByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return cf.GetArrayStringValue(), nil
}

func GetJsonConfigByIDAndSetToObject(ctx context.Context, id int, obj any) (err error) {
	cf, err := GetConfigByID(ctx, id)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(cf.Value), obj)
	if err != nil {
		return fmt.Errorf("[%s] config value is not json format", cf.Key)
	}
	return nil
}

// GetIDByKey get config id by key
func GetIDByKey(ctx context.Context, key string) (id int, err error) {
	cf, err := GetConfigByKey(ctx, key)
	if err != nil {
		return 0, err
	}
	return cf.ID, nil
}

// GenUniqueIDStr generate unique id string
// 1 + 00x(objectType) + 000000000000x(id)
func GenUniqueIDStr(ctx context.Context, key string) (uniqueID string, err error) {
	objectType := constant.ObjectTypeStrMapping[key]
	bean := &entity.Uniqid{UniqidType: objectType}
	_, err = handler.Engine.Context(ctx).Insert(bean)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("1%03d%013d", objectType, bean.ID), nil
}
