package plugin_common

import (
	"context"
	"encoding/json"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/plugin"
	"github.com/lawyer/repo"
	"github.com/lawyer/repo/search_sync"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type PluginConfigRepo interface {
	SavePluginConfig(ctx context.Context, pluginSlugName, configValue string) (err error)
	GetPluginConfigAll(ctx context.Context) (pluginConfigs []*entity.PluginConfig, err error)
}

// PluginCommonService user service
type PluginCommonService struct {
}

// NewPluginCommonService new report service
func NewPluginCommonService() *PluginCommonService {

	// init plugin status
	pluginStatus, err := utils.GetStringValue(context.TODO(), constant.PluginStatus)
	if err != nil {
		log.Error(err)
	} else {
		if err := plugin.StatusManager.UnmarshalJSON([]byte(pluginStatus)); err != nil {
			log.Error(err)
		}
	}

	// init plugin config
	pluginConfigs, err := repo.PluginConfigRepo.GetPluginConfigAll(context.Background())
	if err != nil {
		log.Error(err)
	} else {
		for _, pluginConfig := range pluginConfigs {
			err := plugin.CallConfig(func(fn plugin.Config) error {
				if fn.Info().SlugName == pluginConfig.PluginSlugName {
					return fn.ConfigReceiver([]byte(pluginConfig.Value))
				}
				return nil
			})
			if err != nil {
				log.Errorf("parse plugin config failed: %s %v", pluginConfig.PluginSlugName, err)
			}
		}
	}

	return &PluginCommonService{}
}

// UpdatePluginStatus update plugin status
func (ps *PluginCommonService) UpdatePluginStatus(ctx context.Context) (err error) {
	content, err := plugin.StatusManager.MarshalJSON()
	if err != nil {
		return errors.InternalServer(reason.UnknownError).WithError(err)
	}
	return utils.UpdateConfig(ctx, constant.PluginStatus, string(content))
}

// UpdatePluginConfig update plugin config
func (ps *PluginCommonService) UpdatePluginConfig(ctx context.Context, req *schema.UpdatePluginConfigReq) (err error) {
	configValue, _ := json.Marshal(req.ConfigFields)
	err = repo.PluginConfigRepo.SavePluginConfig(ctx, req.PluginSlugName, string(configValue))
	if err != nil {
		return err
	}

	_ = plugin.CallSearch(func(search plugin.Search) error {
		if search.Info().SlugName == req.PluginSlugName {
			search.RegisterSyncer(ctx, search_sync.NewPluginSyncer(handler.Engine, handler.RedisClient))
		}
		return nil
	})
	return nil
}
