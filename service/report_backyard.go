package service

import (
	"context"
	"encoding/json"
	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/htmltext"
	"github.com/lawyer/repo"
	"github.com/segmentfault/pacman/errors"
)

// ReportAdminServicer user service
type ReportAdminService struct {
}

// NewReportAdminService new report service
func NewReportAdminService() *ReportAdminService {
	return &ReportAdminService{}
}

// ListReportPage list report pages
func (rs *ReportAdminService) ListReportPage(ctx context.Context, dto schema.GetReportListPageDTO) (pageModel *pager.PageModel, err error) {
	var (
		resp  []*schema.GetReportListPageResp
		flags []entity.Report
		total int64

		flaggedUserIds,
		userIds []string

		flaggedUsers,
		users map[string]*schema.UserBasicInfo
	)

	flags, total, err = repo.ReportRepo.GetReportListPage(ctx, dto)
	if err != nil {
		return
	}

	_ = copier.Copy(&resp, flags)
	for _, r := range resp {
		flaggedUserIds = append(flaggedUserIds, r.ReportedUserID)
		userIds = append(userIds, r.UserID)
		r.Format()
	}

	// flagged users
	flaggedUsers, err = UserCommonServicer.BatchUserBasicInfoByID(ctx, flaggedUserIds)
	if err != nil {
		return nil, err
	}

	// flag users
	users, err = UserCommonServicer.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return nil, err
	}
	for _, r := range resp {
		r.ReportedUser = flaggedUsers[r.ReportedUserID]
		r.ReportUser = users[r.UserID]
		rs.decorateReportResp(ctx, r)
	}
	return pager.NewPageModel(total, resp), nil
}

// HandleReported handle the reported object
func (rs *ReportAdminService) HandleReported(ctx context.Context, req schema.ReportHandleReq) (err error) {
	var (
		reported   *entity.Report
		handleData = entity.Report{
			FlaggedContent: req.FlaggedContent,
			FlaggedType:    req.FlaggedType,
			Status:         entity.ReportStatusCompleted,
		}
		exist bool
	)

	reported, exist, err = repo.ReportRepo.GetByID(ctx, req.ID)
	if err != nil {
		err = errors.BadRequest(reason.ReportHandleFailed).WithError(err).WithStack()
		return
	}
	if !exist {
		err = errors.NotFound(reason.ReportNotFound)
		return
	}

	// check if handle or not
	if reported.Status != entity.ReportStatusPending {
		return
	}

	if err = ReportHandler.HandleObject(ctx, reported, req); err != nil {
		return
	}

	err = repo.ReportRepo.UpdateByID(ctx, reported.ID, handleData)
	return
}

func (rs *ReportAdminService) decorateReportResp(ctx context.Context, resp *schema.GetReportListPageResp) {
	lang := utils.GetLangByCtx(ctx)
	objectInfo, err := ObjServicer.GetInfo(ctx, resp.ObjectID)
	if err != nil {
		glog.Slog.Error(err)
		return
	}

	resp.QuestionID = objectInfo.QuestionID
	resp.AnswerID = objectInfo.AnswerID
	resp.CommentID = objectInfo.CommentID
	resp.Title = objectInfo.Title
	resp.Excerpt = htmltext.FetchExcerpt(objectInfo.Content, "...", 240)

	if resp.ReportType > 0 {
		resp.Reason = &schema.ReasonItem{ReasonType: resp.ReportType}
		cf, err := utils.GetConfigByID(ctx, resp.ReportType)
		if err != nil {
			glog.Slog.Error(err)
		} else {
			_ = json.Unmarshal([]byte(cf.Value), resp.Reason)
			resp.Reason.Translate(cf.Key, lang)
		}
	}
	if resp.FlaggedType > 0 {
		resp.FlaggedReason = &schema.ReasonItem{ReasonType: resp.FlaggedType}
		cf, err := utils.GetConfigByID(ctx, resp.FlaggedType)
		if err != nil {
			glog.Slog.Error(err)
		} else {
			_ = json.Unmarshal([]byte(cf.Value), resp.Reason)
			resp.Reason.Translate(cf.Key, lang)
		}
	}
}
