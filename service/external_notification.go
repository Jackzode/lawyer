package service

import (
	"context"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/schema"
)

type ExternalNotificationService struct {
}

func NewExternalNotificationService() *ExternalNotificationService {
	n := &ExternalNotificationService{}
	ExternalNotificationQueueService.RegisterHandler(n.Handler)
	return n
}

func (ns *ExternalNotificationService) Handler(ctx context.Context, msg *schema.ExternalNotificationMsg) error {
	glog.Slog.Debugf("try to send external notification %+v", msg)

	if msg.NewQuestionTemplateRawData != nil {
		return ns.handleNewQuestionNotification(ctx, msg)
	}
	if msg.NewCommentTemplateRawData != nil {
		return ns.handleNewCommentNotification(ctx, msg)
	}
	if msg.NewAnswerTemplateRawData != nil {
		return ns.handleNewAnswerNotification(ctx, msg)
	}
	if msg.NewInviteAnswerTemplateRawData != nil {
		return ns.handleInviteAnswerNotification(ctx, msg)
	}
	glog.Slog.Errorf("unknown notification message: %+v", msg)
	return nil
}
