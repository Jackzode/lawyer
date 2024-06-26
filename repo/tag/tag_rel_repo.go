package tag

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/pkg/uid"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// TagRelRepo tag rel repository
type TagRelRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewTagRelRepo new repository
func NewTagRelRepo() *TagRelRepo {
	return &TagRelRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddTagRelList add tag list
func (tr *TagRelRepo) AddTagRelList(ctx context.Context, tagList []*entity.TagRel) (err error) {
	for _, item := range tagList {
		item.ObjectID = uid.DeShortID(item.ObjectID)
	}
	_, err = tr.DB.Context(ctx).Insert(tagList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if utils.GetEnableShortID(ctx) {
		for _, item := range tagList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// RemoveTagRelListByObjectID delete tag list
func (tr *TagRelRepo) RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.DB.Context(ctx).Where("object_id = ?", objectID).Update(&entity.TagRel{Status: entity.TagRelStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RecoverTagRelListByObjectID recover tag list
func (tr *TagRelRepo) RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.DB.Context(ctx).Where("object_id = ?", objectID).Update(&entity.TagRel{Status: entity.TagRelStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *TagRelRepo) HideTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.DB.Context(ctx).Where("object_id = ?", objectID).Cols("status").Update(&entity.TagRel{Status: entity.TagRelStatusHide})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *TagRelRepo) ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.DB.Context(ctx).Where("object_id = ?", objectID).Cols("status").Update(&entity.TagRel{Status: entity.TagRelStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveTagRelListByIDs delete tag list
func (tr *TagRelRepo) RemoveTagRelListByIDs(ctx context.Context, ids []int64) (err error) {
	_, err = tr.DB.Context(ctx).In("id", ids).Update(&entity.TagRel{Status: entity.TagRelStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetObjectTagRelWithoutStatus get object tag relation no matter status
func (tr *TagRelRepo) GetObjectTagRelWithoutStatus(ctx context.Context, objectID, tagID string) (
	tagRel *entity.TagRel, exist bool, err error,
) {
	objectID = uid.DeShortID(objectID)
	tagRel = &entity.TagRel{}
	session := tr.DB.Context(ctx).Where("object_id = ?", objectID).And("tag_id = ?", tagID)
	exist, err = session.Get(tagRel)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if utils.GetEnableShortID(ctx) {
		tagRel.ObjectID = uid.EnShortID(tagRel.ObjectID)
	}
	return
}

// EnableTagRelByIDs update tag status to available
func (tr *TagRelRepo) EnableTagRelByIDs(ctx context.Context, ids []int64) (err error) {
	_, err = tr.DB.Context(ctx).In("id", ids).Update(&entity.TagRel{Status: entity.TagRelStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetObjectTagRelList get object tag relation list all
func (tr *TagRelRepo) GetObjectTagRelList(ctx context.Context, objectID string) (tagListList []*entity.TagRel, err error) {
	objectID = uid.DeShortID(objectID)
	tagListList = make([]*entity.TagRel, 0)
	session := tr.DB.Context(ctx).Where("object_id = ?", objectID)
	session.In("status", []int{entity.TagRelStatusAvailable, entity.TagRelStatusHide})
	err = session.Find(&tagListList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if utils.GetEnableShortID(ctx) {
		for _, item := range tagListList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// BatchGetObjectTagRelList get object tag relation list all
func (tr *TagRelRepo) BatchGetObjectTagRelList(ctx context.Context, objectIds []string) (tagListList []*entity.TagRel, err error) {
	for num, item := range objectIds {
		objectIds[num] = uid.DeShortID(item)
	}
	tagListList = make([]*entity.TagRel, 0)
	session := tr.DB.Context(ctx).In("object_id", objectIds)
	session.Where("status = ?", entity.TagRelStatusAvailable)
	err = session.Find(&tagListList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if utils.GetEnableShortID(ctx) {
		for _, item := range tagListList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// CountTagRelByTagID count tag relation
func (tr *TagRelRepo) CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error) {
	count, err = tr.DB.Context(ctx).Count(&entity.TagRel{TagID: tagID, Status: entity.AnswerStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
