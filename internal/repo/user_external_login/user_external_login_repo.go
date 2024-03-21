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

package user_external_login

import (
	"context"
	"encoding/json"
	"github.com/apache/incubator-answer/commons/constant"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/commons/entity"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/apache/incubator-answer/commons/schema"
	"github.com/apache/incubator-answer/internal/service/user_external_login"
	"github.com/segmentfault/pacman/errors"
)

type userExternalLoginRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewUserExternalLoginRepo new repository
func NewUserExternalLoginRepo(DB *xorm.Engine, Cache *redis.Client) user_external_login.UserExternalLoginRepo {
	return &userExternalLoginRepo{
		DB:    DB,
		Cache: Cache,
	}
}

// AddUserExternalLogin add external login information
func (ur *userExternalLoginRepo) AddUserExternalLogin(ctx context.Context, user *entity.UserExternalLogin) (err error) {
	_, err = ur.DB.Context(ctx).Insert(user)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateInfo update user info
func (ur *userExternalLoginRepo) UpdateInfo(ctx context.Context, userInfo *entity.UserExternalLogin) (err error) {
	_, err = ur.DB.Context(ctx).ID(userInfo.ID).Update(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetByExternalID get by external ID
func (ur *userExternalLoginRepo) GetByExternalID(ctx context.Context, provider, externalID string) (
	userInfo *entity.UserExternalLogin, exist bool, err error) {
	userInfo = &entity.UserExternalLogin{}
	exist, err = ur.DB.Context(ctx).Where("external_id = ?", externalID).Where("provider = ?", provider).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUserExternalLoginList get by external ID
func (ur *userExternalLoginRepo) GetUserExternalLoginList(ctx context.Context, userID string) (
	resp []*entity.UserExternalLogin, err error) {
	resp = make([]*entity.UserExternalLogin, 0)
	err = ur.DB.Context(ctx).Where("user_id = ?", userID).Find(&resp)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// DeleteUserExternalLogin delete external user login info
func (ur *userExternalLoginRepo) DeleteUserExternalLogin(ctx context.Context, userID, externalID string) (err error) {
	cond := &entity.UserExternalLogin{}
	_, err = ur.DB.Context(ctx).Where("user_id = ? AND external_id = ?", userID, externalID).Delete(cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SetCacheUserExternalLoginInfo cache user info for external login
func (ur *userExternalLoginRepo) SetCacheUserExternalLoginInfo(
	ctx context.Context, key string, info *schema.ExternalLoginUserInfoCache) (err error) {
	cacheData, _ := json.Marshal(info)
	return ur.Cache.Set(ctx, constant.ConnectorUserExternalInfoCacheKey+key,
		string(cacheData), constant.ConnectorUserExternalInfoCacheTime).Err()
}

// GetCacheUserExternalLoginInfo cache user info for external login
func (ur *userExternalLoginRepo) GetCacheUserExternalLoginInfo(
	ctx context.Context, key string) (info *schema.ExternalLoginUserInfoCache, err error) {
	res := ur.Cache.Get(ctx, constant.ConnectorUserExternalInfoCacheKey+key).String()
	if res == "" {
		return nil, nil
	}
	info = &schema.ExternalLoginUserInfoCache{}
	_ = json.Unmarshal([]byte(res), &info)
	return info, nil
}
