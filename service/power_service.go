package service

import (
	"context"
	"github.com/lawyer/commons/entity"
)

// PowerRepo power repository
type PowerRepo interface {
	GetPowerList(ctx context.Context, power *entity.Power) (powers []*entity.Power, err error)
}
