package plugin_config

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
)

type PluginConfigRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewPluginConfigRepo new repository
func NewPluginConfigRepo() *PluginConfigRepo {
	return &PluginConfigRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (ur *PluginConfigRepo) SavePluginConfig(ctx context.Context, pluginSlugName, configValue string) (err error) {
	old := &entity.PluginConfig{PluginSlugName: pluginSlugName}
	exist, err := ur.DB.Context(ctx).Get(old)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		old.Value = configValue
		_, err = ur.DB.Context(ctx).ID(old.ID).Update(old)
	} else {
		_, err = ur.DB.Context(ctx).InsertOne(&entity.PluginConfig{PluginSlugName: pluginSlugName, Value: configValue})
	}
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *PluginConfigRepo) GetPluginConfigAll(ctx context.Context) (pluginConfigs []*entity.PluginConfig, err error) {
	pluginConfigs = make([]*entity.PluginConfig, 0)
	err = ur.DB.Context(ctx).Find(&pluginConfigs)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return pluginConfigs, err
}
