package tag

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/service/tag_common"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// tagRepo tag repository
type tagRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewTagRepo new repository
func NewTagRepo() tag_common.TagRepo {
	return &tagRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// RemoveTag delete tag
func (tr *tagRepo) RemoveTag(ctx context.Context, tagID string) (err error) {
	session := tr.DB.Context(ctx).Where(builder.Eq{"id": tagID})
	_, err = session.Update(&entity.Tag{Status: entity.TagStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateTag update tag
func (tr *tagRepo) UpdateTag(ctx context.Context, tag *entity.Tag) (err error) {
	_, err = tr.DB.Context(ctx).Where(builder.Eq{"id": tag.ID}).Update(tag)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RecoverTag recover deleted tag
func (tr *tagRepo) RecoverTag(ctx context.Context, tagID string) (err error) {
	_, err = tr.DB.Context(ctx).ID(tagID).Update(&entity.Tag{Status: entity.TagStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// MustGetTagByID get tag by id
func (tr *tagRepo) MustGetTagByNameOrID(ctx context.Context, tagID, slugName string) (
	tag *entity.Tag, exist bool, err error) {
	if len(tagID) == 0 && len(slugName) == 0 {
		return nil, false, nil
	}
	tag = &entity.Tag{}
	session := tr.DB.Context(ctx)
	if len(tagID) > 0 {
		session.ID(tagID)
	}
	if len(slugName) > 0 {
		session.Where(builder.Eq{"slug_name": slugName})
	}
	exist, err = session.Get(tag)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateTagSynonym update synonym tag
func (tr *tagRepo) UpdateTagSynonym(ctx context.Context, tagSlugNameList []string, mainTagID int64,
	mainTagSlugName string,
) (err error) {
	bean := &entity.Tag{MainTagID: mainTagID, MainTagSlugName: mainTagSlugName}
	session := tr.DB.Context(ctx).In("slug_name", tagSlugNameList).MustCols("main_tag_id", "main_tag_slug_name")
	_, err = session.Update(bean)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagRepo) GetTagSynonymCount(ctx context.Context, tagID string) (count int64, err error) {
	count, err = tr.DB.Context(ctx).Count(&entity.Tag{MainTagID: converter.StringToInt64(tagID), Status: entity.TagStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetTagList get tag list all
func (tr *tagRepo) GetTagList(ctx context.Context, tag *entity.Tag) (tagList []*entity.Tag, err error) {
	tagList = make([]*entity.Tag, 0)
	session := tr.DB.Context(ctx).Where(builder.Eq{"status": entity.TagStatusAvailable})
	err = session.Find(&tagList, tag)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}