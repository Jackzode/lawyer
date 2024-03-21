package config

import (
	"context"
	"github.com/apache/incubator-answer/commons/entity"
)

// ConfigRepo config repository
type ConfigRepo interface {
	GetConfigByID(ctx context.Context, id int) (c *entity.Config, err error)
	GetConfigByKey(ctx context.Context, key string) (c *entity.Config, err error)
	UpdateConfig(ctx context.Context, key, value string) (err error)
}

// ConfigService user service
type ConfigService struct {
	configRepo ConfigRepo
}

// NewConfigService new config service
func NewConfigService(configRepo ConfigRepo) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
	}
}

// GetConfigByID get config by id
func (cs *ConfigService) GetConfigByID(ctx context.Context, id int) (c *entity.Config, err error) {
	return cs.configRepo.GetConfigByID(ctx, id)
}

func (cs *ConfigService) GetConfigByKey(ctx context.Context, key string) (c *entity.Config, err error) {
	return cs.configRepo.GetConfigByKey(ctx, key)
}

func (cs *ConfigService) UpdateConfig(ctx context.Context, key, value string) (err error) {
	return cs.configRepo.UpdateConfig(ctx, key, value)
}
