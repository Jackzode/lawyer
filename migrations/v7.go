package migrations

import (
	"context"
	"fmt"
	entity2 "github.com/lawyer/commons/entity"

	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

func addPlugin(ctx context.Context, x *xorm.Engine) error {
	defaultConfigTable := []*entity2.Config{
		{ID: 118, Key: "plugin.status", Value: `{}`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Context(ctx).Get(&entity2.Config{ID: c.ID, Key: c.Key})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			continue
		}
		if _, err = x.Context(ctx).Insert(&entity2.Config{ID: c.ID, Key: c.Key, Value: c.Value}); err != nil {
			log.Errorf("insert %+v config failed: %s", c, err)
			return fmt.Errorf("add config failed: %w", err)
		}
	}

	return x.Context(ctx).Sync(new(entity2.PluginConfig), new(entity2.UserExternalLogin))
}
