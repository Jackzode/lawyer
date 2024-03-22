package meta

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	repo "github.com/lawyer/initServer/initRepo"

	"github.com/segmentfault/pacman/errors"
)

// MetaRepo meta repository
type MetaRepo interface {
	AddMeta(ctx context.Context, meta *entity.Meta) (err error)
	RemoveMeta(ctx context.Context, id int) (err error)
	UpdateMeta(ctx context.Context, meta *entity.Meta) (err error)
	GetMetaByObjectIdAndKey(ctx context.Context, objectId, key string) (meta *entity.Meta, exist bool, err error)
	GetMetaList(ctx context.Context, meta *entity.Meta) (metas []*entity.Meta, err error)
}

// MetaService user service
type MetaService struct {
}

func NewMetaService() *MetaService {
	return &MetaService{}
}

// AddMeta add meta
func (ms *MetaService) AddMeta(ctx context.Context, objID, key, value string) (err error) {
	meta := &entity.Meta{
		ObjectID: objID,
		Key:      key,
		Value:    value,
	}
	return repo.MetaRepo.AddMeta(ctx, meta)
}

// RemoveMeta delete meta
func (ms *MetaService) RemoveMeta(ctx context.Context, id int) (err error) {
	return repo.MetaRepo.RemoveMeta(ctx, id)
}

// UpdateMeta update meta
func (ms *MetaService) UpdateMeta(ctx context.Context, metaID int, key, value string) (err error) {
	meta := &entity.Meta{
		ID:    metaID,
		Key:   key,
		Value: value,
	}
	return repo.MetaRepo.UpdateMeta(ctx, meta)
}

// GetMetaByObjectIdAndKey get meta one
func (ms *MetaService) GetMetaByObjectIdAndKey(ctx context.Context, objectID, key string) (meta *entity.Meta, err error) {
	meta, exist, err := repo.MetaRepo.GetMetaByObjectIdAndKey(ctx, objectID, key)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.BadRequest(reason.UnknownError)
	}
	return meta, nil
}

// GetMetaList get meta list all
func (ms *MetaService) GetMetaList(ctx context.Context, objID string) (metas []*entity.Meta, err error) {
	metas, err = repo.MetaRepo.GetMetaList(ctx, &entity.Meta{ObjectID: objID})
	if err != nil {
		return nil, err
	}
	return metas, err
}
