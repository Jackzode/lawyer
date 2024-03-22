package migrations

import (
	"context"
	"github.com/lawyer/commons/entity"
	"xorm.io/xorm"
)

func addNoticeConfig(ctx context.Context, x *xorm.Engine) error {
	return x.Context(ctx).Sync(new(entity.UserNotificationConfig))
}
