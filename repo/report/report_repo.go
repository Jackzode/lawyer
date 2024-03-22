package report

import (
	"context"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/service/report_common"

	"github.com/segmentfault/pacman/errors"
)

// reportRepo report repository
type reportRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewReportRepo new repository
func NewReportRepo() report_common.ReportRepo {
	return &reportRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddReport add report
func (rr *reportRepo) AddReport(ctx context.Context, report *entity.Report) (err error) {
	report.ID, err = repo.UniqueIDRepo.GenUniqueIDStr(ctx, report.TableName())
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
