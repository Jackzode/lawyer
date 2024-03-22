package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"xorm.io/xorm"
)

func addPasswordLoginControl(ctx context.Context, x *xorm.Engine) error {
	loginSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeLogin,
	}
	exist, err := x.Context(ctx).Get(loginSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if exist {
		content := &schema.SiteLoginReq{}
		_ = json.Unmarshal([]byte(loginSiteInfo.Content), content)
		content.AllowPasswordLogin = true
		data, _ := json.Marshal(content)
		loginSiteInfo.Content = string(data)
		_, err = x.Context(ctx).ID(loginSiteInfo.ID).Cols("content").Update(loginSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	}

	writeSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeWrite,
	}
	exist, err = x.Context(ctx).Get(writeSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if exist {
		content := &schema.SiteWriteReq{}
		_ = json.Unmarshal([]byte(writeSiteInfo.Content), content)
		content.RestrictAnswer = true
		data, _ := json.Marshal(content)
		writeSiteInfo.Content = string(data)
		_, err = x.Context(ctx).ID(writeSiteInfo.ID).Cols("content").Update(writeSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	}

	type User struct {
		Avatar string `xorm:"not null default '' VARCHAR(1024) avatar"`
	}
	return x.Context(ctx).Sync(new(User))
}
