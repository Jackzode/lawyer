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

package report

import (
	"context"
	"github.com/apache/incubator-answer/commons/constant"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/commons/entity"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/apache/incubator-answer/commons/schema"
	"github.com/apache/incubator-answer/commons/utils/pager"
	"github.com/apache/incubator-answer/internal/service/report_common"

	"github.com/apache/incubator-answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
)

// reportRepo report repository
type reportRepo struct {
	DB           *xorm.Engine
	Cache        *redis.Client
	uniqueIDRepo unique.UniqueIDRepo
}

// NewReportRepo new repository
func NewReportRepo(DB *xorm.Engine, Cache *redis.Client, uniqueIDRepo unique.UniqueIDRepo) report_common.ReportRepo {
	return &reportRepo{
		DB:           DB,
		Cache:        Cache,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AddReport add report
func (rr *reportRepo) AddReport(ctx context.Context, report *entity.Report) (err error) {
	report.ID, err = rr.uniqueIDRepo.GenUniqueIDStr(ctx, report.TableName())
	if err != nil {
		return err
	}
	_, err = rr.DB.Context(ctx).Insert(report)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetReportListPage get report list page
func (rr *reportRepo) GetReportListPage(ctx context.Context, dto schema.GetReportListPageDTO) (reports []entity.Report, total int64, err error) {
	var (
		ok         bool
		status     int
		objectType int
		session    = rr.DB.Context(ctx)
		cond       = entity.Report{}
	)

	// parse status
	status, ok = entity.ReportStatus[dto.Status]
	if !ok {
		status = entity.ReportStatus["pending"]
	}
	cond.Status = status

	// parse object type
	objectType, ok = constant.ObjectTypeStrMapping[dto.ObjectType]
	if ok {
		cond.ObjectType = objectType
	}

	// order
	session.OrderBy("updated_at desc")

	total, err = pager.Help(dto.Page, dto.PageSize, &reports, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetByID get report by ID
func (rr *reportRepo) GetByID(ctx context.Context, id string) (report *entity.Report, exist bool, err error) {
	report = &entity.Report{}
	exist, err = rr.DB.Context(ctx).ID(id).Get(report)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateByID handle report by ID
func (rr *reportRepo) UpdateByID(ctx context.Context, id string, handleData entity.Report) (err error) {
	_, err = rr.DB.Context(ctx).ID(id).Update(&handleData)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (rr *reportRepo) GetReportCount(ctx context.Context) (count int64, err error) {
	list := make([]*entity.Report, 0)
	count, err = rr.DB.Context(ctx).Where("status =?", entity.ReportStatusPending).FindAndCount(&list)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
