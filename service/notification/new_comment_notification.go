package notification

import (
	"context"
	constant2 "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/repo"
	"github.com/segmentfault/pacman/i18n"
	"github.com/segmentfault/pacman/log"
	"time"
)

func (ns *ExternalNotificationService) handleNewCommentNotification(ctx context.Context,
	msg *schema.ExternalNotificationMsg) error {
	log.Debugf("try to send new comment notification %+v", msg)

	notificationConfig, exist, err := repo.UserNotificationConfigRepo.GetByUserIDAndSource(ctx, msg.ReceiverUserID, constant2.InboxSource)
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
		case constant2.EmailChannel:
			ns.sendNewCommentNotificationEmail(ctx, msg.ReceiverUserID, msg.ReceiverEmail, msg.ReceiverLang, msg.NewCommentTemplateRawData)
		}
	}

	return nil
}

func (ns *ExternalNotificationService) sendNewCommentNotificationEmail(ctx context.Context,
	userID, email, lang string, rawData *schema.NewCommentTemplateRawData) {
	codeContent := &schema.EmailCodeContent{
		SourceType: schema.UnsubscribeSourceType,
		NotificationSources: []constant2.NotificationSource{
			constant2.InboxSource,
		},
		Email:  email,
		UserID: userID,
	}
	// If receiver has set language, use it to send email.
	if len(lang) > 0 {
		ctx = context.WithValue(ctx, constant2.AcceptLanguageFlag, i18n.Language(lang))
	}
	title, body, err := services.EmailService.NewCommentTemplate(ctx, rawData)
	if err != nil {
		log.Error(err)
		return
	}

	services.EmailService.SendAndSaveCodeWithTime(
		ctx, email, title, body, rawData.UnsubscribeCode, codeContent.ToJSONString(), 1*24*time.Hour)
}
