package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/base/translator"
	constant "github.com/lawyer/commons/constant"
	entity "github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/uid"
	"github.com/lawyer/repo"
	"github.com/segmentfault/pacman/log"
)

// NotificationServicer user service
type NotificationService struct {
}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (ns *NotificationService) GetRedDot(ctx context.Context, req *schema.GetRedDot) (*schema.RedDot, error) {
	redBot := &schema.RedDot{}
	inboxKey := fmt.Sprintf("answer_RedDot_%d_%s", schema.NotificationTypeInbox, req.UserID)
	achievementKey := fmt.Sprintf("answer_RedDot_%d_%s", schema.NotificationTypeAchievement, req.UserID)
	inboxValue, err := handler.RedisClient.Get(ctx, inboxKey).Int64()
	if err != nil {
		redBot.Inbox = 0
	} else {
		redBot.Inbox = inboxValue
	}
	achievementValue, err := handler.RedisClient.Get(ctx, achievementKey).Int64()
	if err != nil {
		redBot.Achievement = 0
	} else {
		redBot.Achievement = achievementValue
	}
	revisionCount := &schema.RevisionSearch{}
	_ = copier.Copy(revisionCount, req)
	if req.CanReviewAnswer || req.CanReviewQuestion || req.CanReviewTag {
		redBot.CanRevision = true
		revisionCountNum, err := RevisionComServicer.GetUnreviewedRevisionCount(ctx, revisionCount)
		if err != nil {
			return redBot, err
		}
		redBot.Revision = revisionCountNum
	}

	return redBot, nil
}

func (ns *NotificationService) ClearRedDot(ctx context.Context, req *schema.NotificationClearRequest) (*schema.RedDot, error) {
	botType, ok := schema.NotificationType[req.TypeStr]
	if ok {
		key := fmt.Sprintf("answer_RedDot_%d_%s", botType, req.UserID)
		err := handler.RedisClient.Del(ctx, key).Err()
		if err != nil {
			log.Error("ClearRedDot del cache error", err.Error())
		}
	}
	getRedDotreq := &schema.GetRedDot{}
	_ = copier.Copy(getRedDotreq, req)
	return ns.GetRedDot(ctx, getRedDotreq)
}

func (ns *NotificationService) ClearUnRead(ctx context.Context, userID string, botTypeStr string) error {
	botType, ok := schema.NotificationType[botTypeStr]
	if ok {
		err := repo.NotificationRepo.ClearUnRead(ctx, userID, botType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ns *NotificationService) ClearIDUnRead(ctx context.Context, userID string, id string) error {
	notificationInfo, exist, err := repo.NotificationRepo.GetById(ctx, id)
	if err != nil {
		log.Error("notificationRepo.GetById error", err.Error())
		return nil
	}
	if !exist {
		return nil
	}
	if notificationInfo.UserID == userID && notificationInfo.IsRead == schema.NotificationNotRead {
		err := repo.NotificationRepo.ClearIDUnRead(ctx, userID, id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ns *NotificationService) GetNotificationPage(ctx context.Context, searchCond *schema.NotificationSearch) (
	pageModel *pager.PageModel, err error) {
	resp := make([]*schema.NotificationContent, 0)
	searchType, ok := schema.NotificationType[searchCond.TypeStr]
	if !ok {
		return pager.NewPageModel(0, resp), nil
	}
	searchInboxType := schema.NotificationInboxTypeAll
	if searchType == schema.NotificationTypeInbox {
		_, ok = schema.NotificationInboxType[searchCond.InboxTypeStr]
		if ok {
			searchInboxType = schema.NotificationInboxType[searchCond.InboxTypeStr]
		}
	}
	searchCond.Type = searchType
	searchCond.InboxType = searchInboxType
	notifications, total, err := repo.NotificationRepo.GetNotificationPage(ctx, searchCond)
	if err != nil {
		return nil, err
	}
	resp, err = ns.formatNotificationPage(ctx, notifications)
	if err != nil {
		return nil, err
	}
	return pager.NewPageModel(total, resp), nil
}

func (ns *NotificationService) formatNotificationPage(ctx context.Context, notifications []*entity.Notification) (
	resp []*schema.NotificationContent, err error) {
	lang := utils.GetLangByCtx(ctx)
	enableShortID := utils.GetEnableShortID(ctx)
	userIDs := make([]string, 0)
	userMapping := make(map[string]bool)
	for _, notificationInfo := range notifications {
		item := &schema.NotificationContent{}
		if err := json.Unmarshal([]byte(notificationInfo.Content), item); err != nil {
			log.Error("NotificationContent Unmarshal Error", err.Error())
			continue
		}
		// If notification is downvote, the user info is not needed.
		if item.NotificationAction == constant.NotificationDownVotedTheQuestion ||
			item.NotificationAction == constant.NotificationDownVotedTheAnswer {
			item.UserInfo = nil
		}

		item.ID = notificationInfo.ID
		item.NotificationAction = translator.Tr(lang, item.NotificationAction)
		item.UpdateTime = notificationInfo.UpdatedAt.Unix()
		item.IsRead = notificationInfo.IsRead == schema.NotificationRead

		if enableShortID {
			if answerID, ok := item.ObjectInfo.ObjectMap["answer"]; ok {
				if item.ObjectInfo.ObjectID == answerID {
					item.ObjectInfo.ObjectID = uid.EnShortID(item.ObjectInfo.ObjectMap["answer"])
				}
				item.ObjectInfo.ObjectMap["answer"] = uid.EnShortID(item.ObjectInfo.ObjectMap["answer"])
			}
			if questionID, ok := item.ObjectInfo.ObjectMap["question"]; ok {
				if item.ObjectInfo.ObjectID == questionID {
					item.ObjectInfo.ObjectID = uid.EnShortID(item.ObjectInfo.ObjectMap["question"])
				}
				item.ObjectInfo.ObjectMap["question"] = uid.EnShortID(item.ObjectInfo.ObjectMap["question"])
			}
		}

		if item.UserInfo != nil && !userMapping[item.UserInfo.ID] {
			userIDs = append(userIDs, item.UserInfo.ID)
			userMapping[item.UserInfo.ID] = true
		}
		resp = append(resp, item)
	}

	if len(userIDs) == 0 {
		return resp, nil
	}

	users, err := repo.UserRepo.BatchGetByID(ctx, userIDs)
	if err != nil {
		log.Error(err)
		return resp, nil
	}
	userIDMapping := make(map[string]*entity.User, len(users))
	for _, user := range users {
		userIDMapping[user.ID] = user
	}
	for _, item := range resp {
		if item.UserInfo == nil {
			continue
		}
		userInfo, ok := userIDMapping[item.UserInfo.ID]
		if !ok {
			continue
		}
		if userInfo.Status == entity.UserStatusDeleted {
			item.UserInfo = &schema.UserBasicInfo{
				DisplayName: "user" + converter.DeleteUserDisplay(userInfo.ID),
				Status:      constant.UserDeleted,
			}
		}
	}
	return resp, nil
}
