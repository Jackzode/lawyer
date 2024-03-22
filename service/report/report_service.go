package report

import (
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/pkg/obj"
	"golang.org/x/net/context"
)

// ReportService user service
type ReportService struct {
}

// NewReportService new report service
func NewReportService() *ReportService {
	return &ReportService{}
}

// AddReport add report
func (rs *ReportService) AddReport(ctx context.Context, req *schema.AddReportReq) (err error) {
	objectTypeNumber, err := obj.GetObjectTypeNumberByObjectID(req.ObjectID)
	if err != nil {
		return err
	}

	// TODO this reported user id should be get by revision
	objInfo, err := services.ObjService.GetInfo(ctx, req.ObjectID)
	if err != nil {
		return err
	}

	report := &entity.Report{
		UserID:         req.UserID,
		ReportedUserID: objInfo.ObjectCreatorUserID,
		ObjectID:       req.ObjectID,
		ObjectType:     objectTypeNumber,
		ReportType:     req.ReportType,
		Content:        req.Content,
		Status:         entity.ReportStatusPending,
	}
	return repo.ReportRepo.AddReport(ctx, report)
}
