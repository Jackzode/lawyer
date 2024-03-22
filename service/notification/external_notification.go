package notification

import (
	"context"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/initServer/initServices"
	"github.com/segmentfault/pacman/log"
)

type ExternalNotificationService struct {
}

func NewExternalNotificationService() *ExternalNotificationService {
	n := &ExternalNotificationService{}
	services.ExternalNotificationQueueService.RegisterHandler(n.Handler)
	return n
}

func (ns *ExternalNotificationService) Handler(ctx context.Context, msg *schema.ExternalNotificationMsg) error {
	log.Debugf("try to send external notification %+v", msg)

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
	log.Errorf("unknown notification message: %+v", msg)
	return nil
}
