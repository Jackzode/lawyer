package reason_common

import (
	"context"

	"github.com/lawyer/commons/schema"
)

type ReasonRepo interface {
	ListReasons(ctx context.Context, objectType, action string) (resp []*schema.ReasonItem, err error)
}
