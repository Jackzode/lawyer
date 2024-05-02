package service

import (
	"context"
	constant "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/handler"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/token"
	"github.com/lawyer/repo"
	"github.com/segmentfault/pacman/i18n"
	"time"
)

type NewQuestionSubscriber struct {
	UserID   string                      `json:"user_id"`
	Channels schema.NotificationChannels `json:"channels"`
}

func (ns *ExternalNotificationService) handleNewQuestionNotification(ctx context.Context,
	msg *schema.ExternalNotificationMsg) error {
	glog.Slog.Debugf("try to send new question notification %+v", msg)
	subscribers, err := ns.getNewQuestionSubscribers(ctx, msg)
	if err != nil {
		return err
	}
	glog.Slog.Debugf("get subscribers %d for question %s", len(subscribers), msg.NewQuestionTemplateRawData.QuestionID)

	for _, subscriber := range subscribers {
		for _, channel := range subscriber.Channels {
			if !channel.Enable {
				continue
			}
			switch channel.Key {
			case constant.EmailChannel:
				ns.sendNewQuestionNotificationEmail(ctx, subscriber.UserID, &schema.NewQuestionTemplateRawData{
					QuestionTitle:   msg.NewQuestionTemplateRawData.QuestionTitle,
					QuestionID:      msg.NewQuestionTemplateRawData.QuestionID,
					UnsubscribeCode: token.GenerateToken(),
					Tags:            msg.NewQuestionTemplateRawData.Tags,
					TagIDs:          msg.NewQuestionTemplateRawData.TagIDs,
				})
			}
		}
	}
	return nil
}

func (ns *ExternalNotificationService) getNewQuestionSubscribers(ctx context.Context, msg *schema.ExternalNotificationMsg) (
	subscribers []*NewQuestionSubscriber, err error) {
	subscribersMapping := make(map[string]*NewQuestionSubscriber)

	// 1. get all this new question's tags followers
	tagsFollowerIDs := make([]string, 0)
	followerMapping := make(map[string]bool)
	for _, tagID := range msg.NewQuestionTemplateRawData.TagIDs {
		userIDs, err := repo.FollowRepo.GetFollowUserIDs(ctx, tagID)
		if err != nil {
			glog.Slog.Error(err)
			continue
		}
		for _, userID := range userIDs {
			if _, ok := followerMapping[userID]; ok {
				continue
			}
			followerMapping[userID] = true
			tagsFollowerIDs = append(tagsFollowerIDs, userID)
		}
	}
	userNotificationConfigs, err := repo.UserNotificationConfigRepo.GetByUsersAndSource(
		ctx, tagsFollowerIDs, constant.AllNewQuestionForFollowingTagsSource)
	if err != nil {
		return nil, err
	}
	for _, userNotificationConfig := range userNotificationConfigs {
		if _, ok := subscribersMapping[userNotificationConfig.UserID]; ok {
			continue
		}
		subscribersMapping[userNotificationConfig.UserID] = &NewQuestionSubscriber{
			UserID:   userNotificationConfig.UserID,
			Channels: schema.NewNotificationChannelsFormJson(userNotificationConfig.Channels),
		}
	}
	glog.Slog.Debugf("get %d subscribers from tags", len(subscribersMapping))

	// 2. get all new question's followers
	notificationConfigs, err := repo.UserNotificationConfigRepo.GetBySource(ctx, constant.AllNewQuestionSource)
	if err != nil {
		return nil, err
	}
	for _, notificationConfig := range notificationConfigs {
		if _, ok := subscribersMapping[notificationConfig.UserID]; ok {
			continue
		}
		if ns.checkSendNewQuestionNotificationEmailLimit(ctx, notificationConfig.UserID) {
			continue
		}
		subscribersMapping[notificationConfig.UserID] = &NewQuestionSubscriber{
			UserID:   notificationConfig.UserID,
			Channels: schema.NewNotificationChannelsFormJson(notificationConfig.Channels),
		}
	}

	// 3. remove question owner
	delete(subscribersMapping, msg.NewQuestionTemplateRawData.QuestionAuthorUserID)
	for _, subscriber := range subscribersMapping {
		subscribers = append(subscribers, subscriber)
	}
	glog.Slog.Debugf("get %d subscribers from all new question config", len(subscribers))
	return subscribers, nil
}

func (ns *ExternalNotificationService) checkSendNewQuestionNotificationEmailLimit(ctx context.Context, userID string) bool {
	key := constant.NewQuestionNotificationLimitCacheKeyPrefix + userID
	old, err := handler.RedisClient.Get(ctx, key).Int64()
	if err != nil {
		glog.Slog.Error(err)
		return false
	}
	if old >= constant.NewQuestionNotificationLimitMax {
		glog.Slog.Debugf("%s user reach new question notification limit", userID)
		return true
	}

	err = handler.RedisClient.Incr(ctx, key).Err()
	if err != nil {
		glog.Slog.Error(err)
	}
	return false
}

func (ns *ExternalNotificationService) sendNewQuestionNotificationEmail(ctx context.Context,
	userID string, rawData *schema.NewQuestionTemplateRawData) {
	userInfo, exist, err := repo.UserRepo.GetByUserID(ctx, userID)
	if err != nil {
		glog.Slog.Error(err)
		return
	}
	if !exist {
		glog.Slog.Errorf("user %s not exist", userID)
		return
	}
	// If receiver has set language, use it to send email.
	if len(userInfo.Language) > 0 {
		ctx = context.WithValue(ctx, constant.AcceptLanguageFlag, i18n.Language(userInfo.Language))
	}
	title, body, err := EmailServicer.NewQuestionTemplate(ctx, rawData)
	if err != nil {
		glog.Slog.Error(err)
		return
	}

	codeContent := &schema.EmailCodeContent{
		SourceType: schema.UnsubscribeSourceType,
		Email:      userInfo.EMail,
		UserID:     userID,
		NotificationSources: []constant.NotificationSource{
			constant.AllNewQuestionSource,
			constant.AllNewQuestionForFollowingTagsSource,
		},
	}
	EmailServicer.SendAndSaveCodeWithTime(
		ctx, userInfo.EMail, title, body, rawData.UnsubscribeCode, codeContent.ToJSONString(), 1*24*time.Hour)
}
