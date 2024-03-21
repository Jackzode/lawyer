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

package collection

import (
	"context"
	"github.com/apache/incubator-answer/commons/constant"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/commons/entity"
	"github.com/apache/incubator-answer/commons/utils/pager"
	"github.com/apache/incubator-answer/internal/base/handler"
	collectioncommon "github.com/apache/incubator-answer/internal/service/collection_common"
	"github.com/apache/incubator-answer/internal/service/unique"
	"github.com/apache/incubator-answer/pkg/uid"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// collectionRepo collection repository
type collectionRepo struct {
	DB           *xorm.Engine
	Cache        *redis.Client
	uniqueIDRepo unique.UniqueIDRepo
}

// NewCollectionRepo new repository
func NewCollectionRepo(dDB *xorm.Engine, Cache *redis.Client, uniqueIDRepo unique.UniqueIDRepo) collectioncommon.CollectionRepo {
	return &collectionRepo{
		DB:           dDB,
		Cache:        Cache,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AddCollection add collection
func (cr *collectionRepo) AddCollection(ctx context.Context, collection *entity.Collection) (err error) {
	collection.ID, err = cr.uniqueIDRepo.GenUniqueIDStr(ctx, collection.TableName())
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	_, err = cr.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)
		old := &entity.Collection{
			UserID:   collection.UserID,
			ObjectID: collection.ObjectID,
		}
		exist, err := session.ForUpdate().Get(old)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, nil
		}
		_, err = session.Insert(collection)
		if err != nil {
			return nil, err
		}
		return
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// RemoveCollection delete collection
func (cr *collectionRepo) RemoveCollection(ctx context.Context, id string) (err error) {
	_, err = cr.DB.Context(ctx).Where("id = ?", id).Delete(&entity.Collection{})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// UpdateCollection update collection
func (cr *collectionRepo) UpdateCollection(ctx context.Context, collection *entity.Collection, cols []string) (err error) {
	_, err = cr.DB.Context(ctx).ID(collection.ID).Cols(cols...).Update(collection)
	return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
}

// GetCollection get collection one
func (cr *collectionRepo) GetCollection(ctx context.Context, id int) (collection *entity.Collection, exist bool, err error) {
	collection = &entity.Collection{}
	exist, err = cr.DB.Context(ctx).ID(id).Get(collection)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCollectionList get collection list all
func (cr *collectionRepo) GetCollectionList(ctx context.Context, collection *entity.Collection) (collectionList []*entity.Collection, err error) {
	collectionList = make([]*entity.Collection, 0)
	err = cr.DB.Context(ctx).Find(collectionList, collection)
	err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	return
}

// GetOneByObjectIDAndUser get one by object TagID and user
func (cr *collectionRepo) GetOneByObjectIDAndUser(ctx context.Context, userID string, objectID string) (collection *entity.Collection, exist bool, err error) {
	collection = &entity.Collection{}
	exist, err = cr.DB.Context(ctx).Where("user_id = ? and object_id = ?", userID, objectID).Get(collection)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SearchByObjectIDsAndUser search by object IDs and user
func (cr *collectionRepo) SearchByObjectIDsAndUser(ctx context.Context, userID string, objectIDs []string) ([]*entity.Collection, error) {
	collectionList := make([]*entity.Collection, 0)
	err := cr.DB.Context(ctx).Where("user_id = ?", userID).In("object_id", objectIDs).Find(&collectionList)
	if err != nil {
		return collectionList, err
	}
	return collectionList, nil
}

// CountByObjectID count by object TagID
func (cr *collectionRepo) CountByObjectID(ctx context.Context, objectID string) (total int64, err error) {
	collection := &entity.Collection{}
	total, err = cr.DB.Context(ctx).Where("object_id = ?", objectID).Count(collection)
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCollectionPage get collection page
func (cr *collectionRepo) GetCollectionPage(ctx context.Context, page, pageSize int, collection *entity.Collection) (collectionList []*entity.Collection, total int64, err error) {
	collectionList = make([]*entity.Collection, 0)

	session := cr.DB.Context(ctx)
	if collection.UserID != "" && collection.UserID != "0" {
		session = session.Where("user_id = ?", collection.UserID)
	}

	if collection.UserCollectionGroupID != "" && collection.UserCollectionGroupID != "0" {
		session = session.Where("user_collection_group_id = ?", collection.UserCollectionGroupID)
	}
	session = session.OrderBy("update_time desc")

	total, err = pager.Help(page, pageSize, collectionList, collection, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SearchObjectCollected check object is collected or not
func (cr *collectionRepo) SearchObjectCollected(ctx context.Context, userID string, objectIds []string) (map[string]bool, error) {
	for i := 0; i < len(objectIds); i++ {
		objectIds[i] = uid.DeShortID(objectIds[i])
	}

	list, err := cr.SearchByObjectIDsAndUser(ctx, userID, objectIds)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	collectedMap := make(map[string]bool)
	short := handler.GetEnableShortID(ctx)
	for _, item := range list {
		if short {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
		collectedMap[item.ObjectID] = true
	}
	return collectedMap, nil
}

// SearchList
func (cr *collectionRepo) SearchList(ctx context.Context, search *entity.CollectionSearch) ([]*entity.Collection, int64, error) {
	var count int64
	var err error
	rows := make([]*entity.Collection, 0)
	if search.Page > 0 {
		search.Page = search.Page - 1
	} else {
		search.Page = 0
	}
	if search.PageSize == 0 {
		search.PageSize = constant.DefaultPageSize
	}
	offset := search.Page * search.PageSize
	session := cr.DB.Context(ctx).Where("")
	if len(search.UserID) > 0 {
		session = session.And("user_id = ?", search.UserID)
	} else {
		return rows, count, nil
	}
	session = session.Limit(search.PageSize, offset)
	count, err = session.OrderBy("updated_at desc").FindAndCount(&rows)
	if err != nil {
		return rows, count, err
	}
	return rows, count, nil
}
