package meta

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
)

// MetaRepo meta repository
type MetaRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewMetaRepo new repository
func NewMetaRepo() *MetaRepo {
	return &MetaRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddMeta add meta
func (mr *MetaRepo) AddMeta(ctx context.Context, meta *entity.Meta) (err error) {
	_, err = mr.DB.Context(ctx).Insert(meta)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveMeta delete meta
func (mr *MetaRepo) RemoveMeta(ctx context.Context, id int) (err error) {
	_, err = mr.DB.Context(ctx).ID(id).Delete(&entity.Meta{})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateMeta update meta
func (mr *MetaRepo) UpdateMeta(ctx context.Context, meta *entity.Meta) (err error) {
	_, err = mr.DB.Context(ctx).ID(meta.ID).Update(meta)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetMetaByObjectIdAndKey get meta one
func (mr *MetaRepo) GetMetaByObjectIdAndKey(ctx context.Context, objectID, key string) (
	meta *entity.Meta, exist bool, err error) {

	meta = &entity.Meta{}
	exist, err = mr.DB.Context(ctx).Where(builder.Eq{"object_id": objectID}.And(builder.Eq{"`key`": key})).Desc("created_at").Get(meta)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetMetaList get meta list all
func (mr *MetaRepo) GetMetaList(ctx context.Context, meta *entity.Meta) (metaList []*entity.Meta, err error) {
	metaList = make([]*entity.Meta, 0)
	err = mr.DB.Context(ctx).Find(&metaList, meta)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
