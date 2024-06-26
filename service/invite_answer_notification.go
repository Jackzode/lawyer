package service

import (
	"context"
	"github.com/lawyer/commons/constant"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/repo"
	"github.com/segmentfault/pacman/i18n"
	"time"
)

func (ns *ExternalNotificationService) handleInviteAnswerNotification(ctx context.Context,
	msg *schema.ExternalNotificationMsg) error {
	glog.Slog.Debugf("try to send invite answer notification %+v", msg)

	notificationConfig, exist, err := repo.UserNotificationConfigRepo.GetByUserIDAndSource(ctx, msg.ReceiverUserID, constant.InboxSource)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	channels := schema.NewNotificationChannelsFormJson(notificationConfig.Channels)
	for _, channel := range channels {
		if !channel.Enable {
			continue
		}
		switch channel.Key {
		case constant.EmailChannel:
			ns.sendInviteAnswerNotificationEmail(ctx, msg.ReceiverUserID, msg.ReceiverEmail, msg.ReceiverLang, msg.NewInviteAnswerTemplateRawData)
		}
	}
	return nil
}

func (ns *ExternalNotificationService) sendInviteAnswerNotificationEmail(ctx context.Context,
	userID, email, lang string, rawData *schema.NewInviteAnswerTemplateRawData) {
	codeContent := &schema.EmailCodeContent{
		SourceType: schema.UnsubscribeSourceType,
		NotificationSources: []constant.NotificationSource{
			constant.InboxSource,
		},
		Email:  email,
		UserID: userID,
	}

	// If receiver has set language, use it to send email.
	if len(lang) > 0 {
		ctx = context.WithValue(ctx, constant.AcceptLanguageFlag, i18n.Language(lang))
	}
	title, body, err := EmailServicer.NewInviteAnswerTemplate(ctx, rawData)
	if err != nil {
		glog.Slog.Error(err)
		return
	}

	EmailServicer.SendAndSaveCodeWithTime(
		ctx, email, title, body, rawData.UnsubscribeCode, codeContent.ToJSONString(), 1*24*time.Hour)
}
