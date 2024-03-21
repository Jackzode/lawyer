package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-answer/commons/constant"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/commons/entity"
	"github.com/apache/incubator-answer/initServer/data"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

func GetConfigByID(ctx context.Context, id int) (c *entity.Config, err error) {
	cacheKey := fmt.Sprintf("%s%d", constant.ConfigID2KEYCacheKeyPrefix, id)
	cacheData := data.RedisClient.Get(ctx, cacheKey).String()
	if len(cacheData) > 0 {
		c = &entity.Config{}
		c.BuildByJSON([]byte(cacheData))
		if c.ID > 0 {
			return c, nil
		}
	}

	c = &entity.Config{}
	exist, err := data.Engine.Context(ctx).ID(id).Get(c)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil, fmt.Errorf("config not found by id: %d", id)
	}

	// update cache
	if err := data.RedisClient.Set(ctx, cacheKey, c.JsonString(), constant.ConfigCacheTime).Err(); err != nil {
		log.Error(err)
	}
	return c, nil
}

func GetConfigByKey(ctx context.Context, key string) (c *entity.Config, err error) {
	cacheKey := constant.ConfigKEY2ContentCacheKeyPrefix + key
	cacheData := data.RedisClient.Get(ctx, cacheKey).String()
	if len(cacheData) > 0 {
		c = &entity.Config{}
		c.BuildByJSON([]byte(cacheData))
		if c.ID > 0 {
			return c, nil
		}
	}

	c = &entity.Config{Key: key}
	exist, err := data.Engine.Context(ctx).Get(c)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil, fmt.Errorf("config not found by key: %s", key)
	}

	// update cache
	if err := data.RedisClient.Set(ctx, cacheKey, c.JsonString(), constant.ConfigCacheTime).Err(); err != nil {
		log.Error(err)
	}
	return c, nil
}

func UpdateConfig(ctx context.Context, key string, value string) (err error) {
	// check if key exists
	oldConfig := &entity.Config{Key: key}
	exist, err := data.Engine.Context(ctx).Get(oldConfig)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return errors.BadRequest(reason.ObjectNotFound)
	}

	// update database
	_, err = data.Engine.Context(ctx).ID(oldConfig.ID).Update(&entity.Config{Value: value})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	oldConfig.Value = value
	cacheVal := oldConfig.JsonString()
	// update cache
	if err := data.RedisClient.Set(ctx,
		constant.ConfigKEY2ContentCacheKeyPrefix+key, cacheVal, constant.ConfigCacheTime).Err(); err != nil {
		log.Error(err)
	}
	if err := data.RedisClient.Set(ctx,
		fmt.Sprintf("%s%d", constant.ConfigID2KEYCacheKeyPrefix, oldConfig.ID), cacheVal, constant.ConfigCacheTime).Err(); err != nil {
		log.Error(err)
	}
	return
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
