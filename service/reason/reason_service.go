package reason

import (
	"context"
	"github.com/lawyer/commons/schema"
	repo "github.com/lawyer/initServer/initRepo"
)

type ReasonService struct {
}

func NewReasonService() *ReasonService {
	return &ReasonService{}
}

func (rs ReasonService) GetReasons(ctx context.Context, req schema.ReasonReq) (resp []*schema.ReasonItem, err error) {
	return repo.ReasonRepo.ListReasons(ctx, req.ObjectType, req.Action)
}
