package migrations

import (
	"context"
	"fmt"
	entity "github.com/lawyer/commons/entity"

	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

func addPlugin(ctx context.Context, x *xorm.Engine) error {
	defaultConfigTable := []*entity.Config{
		{ID: 118, Key: "plugin.status", Value: `{}`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Context(ctx).Get(&entity.Config{ID: c.ID, Key: c.Key})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			continue
		}
		if _, err = x.Context(ctx).Insert(&entity.Config{ID: c.ID, Key: c.Key, Value: c.Value}); err != nil {
			log.Errorf("insert %+v config failed: %s", c, err)
			return fmt.Errorf("add config failed: %w", err)
		}
	}

	return x.Context(ctx).Sync(new(entity.PluginConfig), new(entity.UserExternalLogin))
}
