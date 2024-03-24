package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/mojocn/base64Captcha"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"image/color"
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
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil, fmt.Errorf("config not found by id: %d", id)
	}

	// update cache
	if err := handler.RedisClient.Set(ctx, cacheKey, c.JsonString(), constant.ConfigCacheTime).Err(); err != nil {
		log.Error(err)
	}
	return c, nil
}

func GetConfigByKey(ctx context.Context, key string) (c *entity.Config, err error) {
	cacheKey := constant.ConfigKEY2ContentCacheKeyPrefix + key
	cachehandler := handler.RedisClient.Get(ctx, cacheKey).String()
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
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil, fmt.Errorf("config not found by key: %s", key)
	}

	// update cache
	if err := handler.RedisClient.Set(ctx, cacheKey, c.JsonString(), constant.ConfigCacheTime).Err(); err != nil {
		log.Error(err)
	}
	return c, nil
}

func UpdateConfig(ctx context.Context, key string, value string) (err error) {
	// check if key exists
	oldConfig := &entity.Config{Key: key}
	exist, err := handler.Engine.Context(ctx).Get(oldConfig)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return errors.BadRequest(reason.ObjectNotFound)
	}

	// update handlerbase
	_, err = handler.Engine.Context(ctx).ID(oldConfig.ID).Update(&entity.Config{Value: value})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	oldConfig.Value = value
	cacheVal := oldConfig.JsonString()
	// update cache
	if err := handler.RedisClient.Set(ctx,
		constant.ConfigKEY2ContentCacheKeyPrefix+key, cacheVal, constant.ConfigCacheTime).Err(); err != nil {
		log.Error(err)
	}
	if err := handler.RedisClient.Set(ctx,
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

func GenerateCaptcha(ctx context.Context) (key, captchaBase64 string, err error) {
	driverString := base64Captcha.DriverString{
		Height:          60,
		Width:           200,
		NoiseCount:      0,
		ShowLineOptions: 2 | 4,
		Length:          4,
		Source:          "1234567890qwertyuioplkjhgfdsazxcvbnm",
		BgColor:         &color.RGBA{R: 211, G: 211, B: 211, A: 0},
		Fonts:           []string{"wqy-microhei.ttc"},
	}
	driver := driverString.ConvertFonts()

	id, content, answer := driver.GenerateIdQuestionAnswer()
	item, err := driver.DrawCaptcha(content)
	if err != nil {
		return "", "", errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	err = repo.CaptchaRepo.SetCaptcha(ctx, id, answer)
	if err != nil {
		return "", "", err
	}

	captchaBase64 = item.EncodeB64string()
	return id, captchaBase64, nil
}
