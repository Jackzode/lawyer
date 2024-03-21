/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package tag

import (
	"context"
	"github.com/apache/incubator-answer/commons/constant/reason"
	entity2 "github.com/apache/incubator-answer/commons/entity"
	"github.com/apache/incubator-answer/internal/base/handler"
	tagcommon "github.com/apache/incubator-answer/internal/service/tag_common"
	"github.com/apache/incubator-answer/internal/service/unique"
	"github.com/apache/incubator-answer/pkg/uid"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// tagRelRepo tag rel repository
type tagRelRepo struct {
	DB           *xorm.Engine
	Cache        *redis.Client
	uniqueIDRepo unique.UniqueIDRepo
}

// NewTagRelRepo new repository
func NewTagRelRepo(DB *xorm.Engine, Cache *redis.Client,
	uniqueIDRepo unique.UniqueIDRepo) tagcommon.TagRelRepo {
	return &tagRelRepo{
		DB:           DB,
		Cache:        Cache,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AddTagRelList add tag list
func (tr *tagRelRepo) AddTagRelList(ctx context.Context, tagList []*entity2.TagRel) (err error) {
	for _, item := range tagList {
		item.ObjectID = uid.DeShortID(item.ObjectID)
	}
	_, err = tr.DB.Context(ctx).Insert(tagList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range tagList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// RemoveTagRelListByObjectID delete tag list
func (tr *tagRelRepo) RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.DB.Context(ctx).Where("object_id = ?", objectID).Update(&entity2.TagRel{Status: entity2.TagRelStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RecoverTagRelListByObjectID recover tag list
func (tr *tagRelRepo) RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.DB.Context(ctx).Where("object_id = ?", objectID).Update(&entity2.TagRel{Status: entity2.TagRelStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagRelRepo) HideTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.DB.Context(ctx).Where("object_id = ?", objectID).Cols("status").Update(&entity2.TagRel{Status: entity2.TagRelStatusHide})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagRelRepo) ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.DB.Context(ctx).Where("object_id = ?", objectID).Cols("status").Update(&entity2.TagRel{Status: entity2.TagRelStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveTagRelListByIDs delete tag list
func (tr *tagRelRepo) RemoveTagRelListByIDs(ctx context.Context, ids []int64) (err error) {
	_, err = tr.DB.Context(ctx).In("id", ids).Update(&entity2.TagRel{Status: entity2.TagRelStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetObjectTagRelWithoutStatus get object tag relation no matter status
func (tr *tagRelRepo) GetObjectTagRelWithoutStatus(ctx context.Context, objectID, tagID string) (
	tagRel *entity2.TagRel, exist bool, err error,
) {
	objectID = uid.DeShortID(objectID)
	tagRel = &entity2.TagRel{}
	session := tr.DB.Context(ctx).Where("object_id = ?", objectID).And("tag_id = ?", tagID)
	exist, err = session.Get(tagRel)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if handler.GetEnableShortID(ctx) {
		tagRel.ObjectID = uid.EnShortID(tagRel.ObjectID)
	}
	return
}

// EnableTagRelByIDs update tag status to available
func (tr *tagRelRepo) EnableTagRelByIDs(ctx context.Context, ids []int64) (err error) {
	_, err = tr.DB.Context(ctx).In("id", ids).Update(&entity2.TagRel{Status: entity2.TagRelStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetObjectTagRelList get object tag relation list all
func (tr *tagRelRepo) GetObjectTagRelList(ctx context.Context, objectID string) (tagListList []*entity2.TagRel, err error) {
	objectID = uid.DeShortID(objectID)
	tagListList = make([]*entity2.TagRel, 0)
	session := tr.DB.Context(ctx).Where("object_id = ?", objectID)
	session.In("status", []int{entity2.TagRelStatusAvailable, entity2.TagRelStatusHide})
	err = session.Find(&tagListList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range tagListList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// BatchGetObjectTagRelList get object tag relation list all
func (tr *tagRelRepo) BatchGetObjectTagRelList(ctx context.Context, objectIds []string) (tagListList []*entity2.TagRel, err error) {
	for num, item := range objectIds {
		objectIds[num] = uid.DeShortID(item)
	}
	tagListList = make([]*entity2.TagRel, 0)
	session := tr.DB.Context(ctx).In("object_id", objectIds)
	session.Where("status = ?", entity2.TagRelStatusAvailable)
	err = session.Find(&tagListList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range tagListList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// CountTagRelByTagID count tag relation
func (tr *tagRelRepo) CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error) {
	count, err = tr.DB.Context(ctx).Count(&entity2.TagRel{TagID: tagID, Status: entity2.AnswerStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
