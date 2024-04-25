package report_handle_admin

import (
	"context"
	constant "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/pkg/obj"
	"github.com/lawyer/repo"
)

type ReportHandle struct {
}

func NewReportHandle() *ReportHandle {
	return &ReportHandle{}
}

// HandleObject this handle object status
func (rh *ReportHandle) HandleObject(ctx context.Context, reported *entity.Report, req schema.ReportHandleReq) (err error) {
	reasonDeleteCfg, err := utils.GetConfigByKey(ctx, "reason.needs_delete")
	if err != nil {
		return err
	}
	reasonCloseCfg, err := utils.GetConfigByKey(ctx, "reason.needs_close")
	if err != nil {
		return err
	}
	var (
		objectID       = reported.ObjectID
		reportedUserID = reported.ReportedUserID
		objectKey      string
	)

	objectKey, err = obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return err
	}
	switch objectKey {
	case "question":
		switch req.FlaggedType {
		case reasonDeleteCfg.ID:
			err = services.QuestionCommon.RemoveQuestion(ctx, &schema.RemoveQuestionReq{ID: objectID})
		case reasonCloseCfg.ID:
			err = services.QuestionCommon.CloseQuestion(ctx, &schema.CloseQuestionReq{
				ID:        objectID,
				CloseType: req.FlaggedType,
				CloseMsg:  req.FlaggedContent,
			})
		}
	case "answer":
		switch req.FlaggedType {
		case reasonDeleteCfg.ID:
			err = services.QuestionCommon.RemoveAnswer(ctx, objectID)
		}
	case "comment":
		switch req.FlaggedType {
		case reasonCloseCfg.ID:
			err = repo.CommentRepo.RemoveComment(ctx, objectID)
			rh.sendNotification(ctx, reportedUserID, objectID, constant.NotificationYourCommentWasDeleted)
		}
	}
	return
}

// sendNotification send rank triggered notification
func (rh *ReportHandle) sendNotification(ctx context.Context, reportedUserID, objectID, notificationAction string) {
	msg := &schema.NotificationMsg{
		TriggerUserID:      reportedUserID,
		ReceiverUserID:     reportedUserID,
		Type:               schema.NotificationTypeInbox,
		ObjectID:           objectID,
		ObjectType:         constant.ReportObjectType,
		NotificationAction: notificationAction,
	}
	services.NotificationQueueService.Send(ctx, msg)
}
