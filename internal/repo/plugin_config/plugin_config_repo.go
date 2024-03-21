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

package plugin_config

import (
	"context"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/commons/entity"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/apache/incubator-answer/internal/service/plugin_common"
	"github.com/segmentfault/pacman/errors"
)

type pluginConfigRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewPluginConfigRepo new repository
func NewPluginConfigRepo(DB *xorm.Engine, Cache *redis.Client) plugin_common.PluginConfigRepo {
	return &pluginConfigRepo{
		DB: DB, Cache: Cache,
	}
}

func (ur *pluginConfigRepo) SavePluginConfig(ctx context.Context, pluginSlugName, configValue string) (err error) {
	old := &entity.PluginConfig{PluginSlugName: pluginSlugName}
	exist, err := ur.DB.Context(ctx).Get(old)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		old.Value = configValue
		_, err = ur.DB.Context(ctx).ID(old.ID).Update(old)
	} else {
		_, err = ur.DB.Context(ctx).InsertOne(&entity.PluginConfig{PluginSlugName: pluginSlugName, Value: configValue})
	}
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *pluginConfigRepo) GetPluginConfigAll(ctx context.Context) (pluginConfigs []*entity.PluginConfig, err error) {
	pluginConfigs = make([]*entity.PluginConfig, 0)
	err = ur.DB.Context(ctx).Find(&pluginConfigs)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return pluginConfigs, err
}
