package reason

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/segmentfault/pacman/log"
)

type ReasonRepo struct {
}

func NewReasonRepo() *ReasonRepo {
	return &ReasonRepo{}
}

func (rr *ReasonRepo) ListReasons(ctx context.Context, objectType, action string) (resp []*schema.ReasonItem, err error) {
	lang := utils.GetLangByCtx(ctx)
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
