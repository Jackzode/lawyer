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

package role

import (
	"context"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/commons/entity"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	service "github.com/apache/incubator-answer/internal/service/role"
	"github.com/segmentfault/pacman/errors"
)

// roleRepo role repository
type roleRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewRoleRepo new repository
func NewRoleRepo(DB *xorm.Engine, Cache *redis.Client) service.RoleRepo {
	return &roleRepo{
		DB:    DB,
		Cache: Cache,
	}
}

// GetRoleAllList get role list all
func (rr *roleRepo) GetRoleAllList(ctx context.Context) (roleList []*entity.Role, err error) {
	roleList = make([]*entity.Role, 0)
	err = rr.DB.Context(ctx).Find(&roleList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetRoleAllMapping get role all mapping
func (rr *roleRepo) GetRoleAllMapping(ctx context.Context) (roleMapping map[int]*entity.Role, err error) {
	roleList, err := rr.GetRoleAllList(ctx)
	if err != nil {
		return nil, err
	}
	roleMapping = make(map[int]*entity.Role, 0)
	for _, role := range roleList {
		roleMapping[role.ID] = role
	}
	return roleMapping, nil
}
