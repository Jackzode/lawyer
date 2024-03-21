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

package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/commons/entity"
	"github.com/redis/go-redis/v9"
	"time"
	"xorm.io/xorm"

	"github.com/apache/incubator-answer/internal/service/action"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// captchaRepo captcha repository
type captchaRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewCaptchaRepo new repository
func NewCaptchaRepo(DB *xorm.Engine, Cache *redis.Client) action.CaptchaRepo {
	return &captchaRepo{
		DB:    DB,
		Cache: Cache,
	}
}

func (cr *captchaRepo) SetActionType(ctx context.Context, unit, actionType, config string, amount int) (err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", unit, actionType, now.Format("2006-1-02"))
	value := &entity.ActionRecordInfo{}
	value.LastTime = now.Unix()
	value.Num = amount
	valueStr, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	err = cr.Cache.Set(ctx, cacheKey, string(valueStr), 6*time.Minute).Err()
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (cr *captchaRepo) GetActionType(ctx context.Context, Ip, actionType string) (actionInfo *entity.ActionRecordInfo, err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", Ip, actionType, now.Format("2006-1-02"))
	res := cr.Cache.Get(ctx, cacheKey).String()
	if res == "" {
		return nil, nil
	}
	actionInfo = &entity.ActionRecordInfo{}
	_ = json.Unmarshal([]byte(res), actionInfo)
	return actionInfo, nil
}

func (cr *captchaRepo) DelActionType(ctx context.Context, unit, actionType string) (err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", unit, actionType, now.Format("2006-1-02"))
	err = cr.Cache.Del(ctx, cacheKey).Err()
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SetCaptcha set captcha to cache
func (cr *captchaRepo) SetCaptcha(ctx context.Context, key, captcha string) (err error) {
	err = cr.Cache.Set(ctx, key, captcha, 6*time.Minute).Err()
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCaptcha get captcha from cache
func (cr *captchaRepo) GetCaptcha(ctx context.Context, key string) (captcha string, err error) {
	captcha = cr.Cache.Get(ctx, key).String()
	if captcha == "" {
		return "", fmt.Errorf("captcha not exist")
	}
	return captcha, nil
}

func (cr *captchaRepo) DelCaptcha(ctx context.Context, key string) (err error) {
	err = cr.Cache.Del(ctx, key).Err()
	if err != nil {
		log.Debug(err)
	}
	return nil
}
