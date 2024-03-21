package reason

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-answer/commons/schema"
	"github.com/apache/incubator-answer/commons/utils"
	"github.com/apache/incubator-answer/internal/base/handler"
	"github.com/apache/incubator-answer/internal/service/reason_common"
	"github.com/segmentfault/pacman/log"
)

type reasonRepo struct {
}

func NewReasonRepo() reason_common.ReasonRepo {
	return &reasonRepo{}
}

func (rr *reasonRepo) ListReasons(ctx context.Context, objectType, action string) (resp []*schema.ReasonItem, err error) {
	lang := handler.GetLangByCtx(ctx)
	reasonAction := fmt.Sprintf("%s.%s.reasons", objectType, action)
	resp = make([]*schema.ReasonItem, 0)

	reasonKeys, err := utils.GetArrayStringValue(ctx, reasonAction)
	if err != nil {
		return nil, err
	}
	for _, reasonKey := range reasonKeys {
		cfg, err := utils.GetConfigByKey(ctx, reasonKey)
		if err != nil {
			log.Error(err)
			continue
		}

		reason := &schema.ReasonItem{}
		err = json.Unmarshal(cfg.GetByteValue(), reason)
		if err != nil {
			log.Error(err)
			continue
		}
		reason.Translate(reasonKey, lang)
		reason.ReasonType = cfg.ID
		resp = append(resp, reason)
	}
	return resp, nil
}
