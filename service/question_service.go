package service

import (
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/base/validator"
	constant "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity "github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/site"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/pkg/token"
	"github.com/lawyer/repo"
	"github.com/lawyer/service/permission"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/htmltext"
	"github.com/lawyer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"golang.org/x/net/context"
)

// QuestionRepo question repository

// QuestionServicer user service
type QuestionService struct {
}

func NewQuestionService() *QuestionService {
	return &QuestionService{}
}

func (qs *QuestionService) CloseQuestion(ctx context.Context, req *schema.CloseQuestionReq) error {
	questionInfo, has, err := repo.QuestionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	questionInfo.Status = entity.QuestionStatusClosed
	err = repo.QuestionRepo.UpdateQuestionStatus(ctx, questionInfo.ID, questionInfo.Status)
	if err != nil {
		return err
	}

	closeMeta, _ := json.Marshal(schema.CloseQuestionMeta{
		CloseType: req.CloseType,
		CloseMsg:  req.CloseMsg,
	})
	err = MetaService.AddMeta(ctx, req.ID, entity.QuestionCloseReasonKey, string(closeMeta))
	if err != nil {
		return err
	}

	ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionClosed,
	})
	return nil
}

// ReopenQuestion reopen question
func (qs *QuestionService) ReopenQuestion(ctx context.Context, req *schema.ReopenQuestionReq) error {
	questionInfo, has, err := repo.QuestionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	questionInfo.Status = entity.QuestionStatusAvailable
	err = repo.QuestionRepo.UpdateQuestionStatus(ctx, questionInfo.ID, questionInfo.Status)
	if err != nil {
		return err
	}
	ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionReopened,
	})
	return nil
}

func (qs *QuestionService) AddQuestionCheckTags(ctx context.Context, Tags []*entity.Tag) ([]string, error) {
	list := make([]string, 0)
	for _, tag := range Tags {
		if tag.Reserved {
			list = append(list, tag.DisplayName)
		}
	}
	if len(list) > 0 {
		return list, errors.BadRequest(reason.RequestFormatError)
	}
	return []string{}, nil
}

func (qs *QuestionService) CheckAddQuestion(ctx context.Context, req *schema.QuestionAdd) (errorlist any, err error) {
	if len(req.Tags) == 0 {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.TagNotFound),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}
	recommendExist, err := TagServicer.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	tagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tagNameList = append(tagNameList, tag.SlugName)
	}
	Tags, tagerr := TagServicer.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return errorlist, tagerr
	}
	if !req.QuestionPermission.CanUseReservedTag {
		taglist, err := qs.AddQuestionCheckTags(ctx, Tags)
		errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
			strings.Join(taglist, ","))
		if err != nil {
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RecommendTagEnter)
			return errorlist, err
		}
	}
	return nil, nil
}

// HasNewTag
func (qs *QuestionService) HasNewTag(ctx context.Context, tags []*schema.TagItem) (bool, error) {
	return TagServicer.HasNewTag(ctx, tags)
}

// AddQuestion add question
func (qs *QuestionService) AddQuestion(ctx context.Context, req *schema.QuestionAdd) (questionInfo any, err error) {
	if len(req.Tags) == 0 {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.TagNotFound),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}
	recommendExist, err := TagServicer.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	tagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tag.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		tagNameList = append(tagNameList, tag.SlugName)
	}
	tags, tagerr := TagServicer.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return questionInfo, tagerr
	}
	if !req.QuestionPermission.CanUseReservedTag {
		taglist, err := qs.AddQuestionCheckTags(ctx, tags)
		errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
			strings.Join(taglist, ","))
		if err != nil {
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RecommendTagEnter)
			return errorlist, err
		}
	}

	question := &entity.Question{}
	now := time.Now()
	question.UserID = req.UserID
	question.Title = req.Title
	question.OriginalText = req.Content
	question.ParsedText = req.HTML
	question.AcceptedAnswerID = "0"
	question.LastAnswerID = "0"
	question.LastEditUserID = "0"
	//question.PostUpdateTime = nil
	question.Status = entity.QuestionStatusAvailable
	question.RevisionID = "0"
	question.CreatedAt = now
	question.PostUpdateTime = now
	question.Pin = entity.QuestionUnPin
	question.Show = entity.QuestionShow
	//question.UpdatedAt = nil
	err = repo.QuestionRepo.AddQuestion(ctx, question)
	if err != nil {
		return
	}
	objectTagData := schema.TagChange{}
	objectTagData.ObjectID = question.ID
	objectTagData.Tags = req.Tags
	objectTagData.UserID = req.UserID
	err = qs.ChangeTag(ctx, &objectTagData)
	if err != nil {
		return
	}

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   question.UserID,
		ObjectID: question.ID,
		Title:    question.Title,
	}

	questionWithTagsRevision, err := qs.changeQuestionToRevision(ctx, question, tags)
	if err != nil {
		return nil, err
	}
	infoJSON, _ := json.Marshal(questionWithTagsRevision)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := RevisionComServicer.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return
	}

	// user add question count
	userQuestionCount, err := QuestionCommonServicer.GetUserQuestionCount(ctx, question.UserID)
	if err != nil {
		glog.Slog.Errorf("get user question count error %v", err)
	} else {
		err = UserCommonServicer.UpdateQuestionCount(ctx, question.UserID, userQuestionCount)
		if err != nil {
			glog.Slog.Errorf("update user question count error %v", err)
		}
	}

	ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
		UserID:           question.UserID,
		ObjectID:         question.ID,
		OriginalObjectID: question.ID,
		ActivityTypeKey:  constant.ActQuestionAsked,
		RevisionID:       revisionID,
	})

	ExternalNotificationQueueService.Send(ctx,
		schema.CreateNewQuestionNotificationMsg(question.ID, question.Title, question.UserID, tags))

	questionInfo, err = qs.GetQuestion(ctx, question.ID, question.UserID, req.QuestionPermission)
	return
}

// OperationQuestion
func (qs *QuestionService) OperationQuestion(ctx context.Context, req *schema.OperationQuestionReq) (err error) {
	questionInfo, has, err := repo.QuestionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}
	// Hidden question cannot be placed at the top
	if questionInfo.Show == entity.QuestionHide && req.Operation == schema.QuestionOperationPin {
		return nil
	}
	// Question cannot be hidden when they are at the top
	if questionInfo.Pin == entity.QuestionPin && req.Operation == schema.QuestionOperationHide {
		return nil
	}

	switch req.Operation {
	case schema.QuestionOperationHide:
		questionInfo.Show = entity.QuestionHide
		err = TagServicer.HideTagRelListByObjectID(ctx, req.ID)
		if err != nil {
			return err
		}
		err = TagServicer.RefreshTagCountByQuestionID(ctx, req.ID)
		if err != nil {
			return err
		}
	case schema.QuestionOperationShow:
		questionInfo.Show = entity.QuestionShow
		err = TagServicer.ShowTagRelListByObjectID(ctx, req.ID)
		if err != nil {
			return err
		}
		err = TagServicer.RefreshTagCountByQuestionID(ctx, req.ID)
		if err != nil {
			return err
		}
	case schema.QuestionOperationPin:
		questionInfo.Pin = entity.QuestionPin
	case schema.QuestionOperationUnPin:
		questionInfo.Pin = entity.QuestionUnPin
	}

	err = repo.QuestionRepo.UpdateQuestionOperation(ctx, questionInfo)
	if err != nil {
		return err
	}

	actMap := make(map[string]constant.ActivityTypeKey)
	actMap[schema.QuestionOperationPin] = constant.ActQuestionPin
	actMap[schema.QuestionOperationUnPin] = constant.ActQuestionUnPin
	actMap[schema.QuestionOperationHide] = constant.ActQuestionHide
	actMap[schema.QuestionOperationShow] = constant.ActQuestionShow
	_, ok := actMap[req.Operation]
	if ok {
		ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  actMap[req.Operation],
		})
	}

	return nil
}

// RemoveQuestion delete question
func (qs *QuestionService) RemoveQuestion(ctx context.Context, req *schema.RemoveQuestionReq) (err error) {
	questionInfo, has, err := repo.QuestionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	//if the status is deleted, return directly
	if questionInfo.Status == entity.QuestionStatusDeleted {
		return nil
	}
	if !has {
		return nil
	}
	if !req.IsAdmin {
		if questionInfo.UserID != req.UserID {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}

		if questionInfo.AcceptedAnswerID != "0" {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}
		if questionInfo.AnswerCount > 1 {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}

		if questionInfo.AnswerCount == 1 {
			answersearch := &entity.AnswerSearch{}
			answersearch.QuestionID = req.ID
			//todo
			answerList, _, err := AnswerCommonServicer.Search(ctx, answersearch)
			if err != nil {
				return err
			}
			for _, answer := range answerList {
				if answer.VoteCount > 0 {
					return errors.BadRequest(reason.QuestionCannotDeleted)
				}
			}
		}
	}

	questionInfo.Status = entity.QuestionStatusDeleted
	err = repo.QuestionRepo.UpdateQuestionStatusWithOutUpdateTime(ctx, questionInfo)
	if err != nil {
		return err
	}

	userQuestionCount, err := QuestionCommonServicer.GetUserQuestionCount(ctx, questionInfo.UserID)
	if err != nil {
		glog.Slog.Error("user GetUserQuestionCount error", err.Error())
	} else {
		err = UserCommonServicer.UpdateQuestionCount(ctx, questionInfo.UserID, userQuestionCount)
		if err != nil {
			glog.Slog.Error("user IncreaseQuestionCount error", err.Error())
		}
	}

	//tag count
	tagIDs := make([]string, 0)
	Tags, tagerr := TagServicer.GetObjectEntityTag(ctx, req.ID)
	if tagerr != nil {
		glog.Slog.Error("GetObjectEntityTag error", tagerr)
		return nil
	}
	for _, v := range Tags {
		tagIDs = append(tagIDs, v.ID)
	}
	err = TagServicer.RemoveTagRelListByObjectID(ctx, req.ID)
	if err != nil {
		glog.Slog.Error("RemoveTagRelListByObjectID error", err.Error())
	}
	err = TagServicer.RefreshTagQuestionCount(ctx, tagIDs)
	if err != nil {
		glog.Slog.Error("efreshTagQuestionCount error", err.Error())
	}

	// #2372 In order to simplify the process and complexity, as well as to consider if it is in-house,
	// facing the problem of recovery.
	// err = qs.answerActivityService.DeleteQuestion(ctx, questionInfo.ID, questionInfo.CreatedAt, questionInfo.VoteCount)
	// if err != nil {
	// 	 glog.Slog.Errorf("user DeleteQuestion rank rollback error %s", err.Error())
	// }
	ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionDeleted,
	})
	return nil
}

func (qs *QuestionService) UpdateQuestionCheckTags(ctx context.Context, req *schema.QuestionUpdate) (errorlist []*validator.FormErrorField, err error) {
	dbinfo, has, err := repo.QuestionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return
	}
	if !has {
		return
	}

	oldTags, tagerr := TagServicer.GetObjectEntityTag(ctx, req.ID)
	if tagerr != nil {
		glog.Slog.Error("GetObjectEntityTag error", tagerr)
		return nil, nil
	}

	tagNameList := make([]string, 0)
	oldtagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tagNameList = append(tagNameList, tag.SlugName)
	}
	for _, tag := range oldTags {
		oldtagNameList = append(oldtagNameList, tag.SlugName)
	}

	isChange := TagServicer.CheckTagsIsChange(ctx, tagNameList, oldtagNameList)

	//If the content is the same, ignore it
	if dbinfo.Title == req.Title && dbinfo.OriginalText == req.Content && !isChange {
		return
	}

	Tags, tagerr := TagServicer.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		glog.Slog.Error("GetTagListByNames error", tagerr)
		return nil, nil
	}

	// if user can not use reserved tag, old reserved tag can not be removed and new reserved tag can not be added.
	if !req.CanUseReservedTag {
		CheckOldTag, CheckNewTag, CheckOldTaglist, CheckNewTaglist := qs.CheckChangeReservedTag(ctx, oldTags, Tags)
		if !CheckOldTag {
			errMsg := fmt.Sprintf(`The reserved tag "%s" must be present.`,
				strings.Join(CheckOldTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
		if !CheckNewTag {
			errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
				strings.Join(CheckNewTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
	}
	return nil, nil
}

func (qs *QuestionService) RecoverQuestion(ctx context.Context, req *schema.QuestionRecoverReq) (err error) {
	questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}
	if questionInfo.Status != entity.QuestionStatusDeleted {
		return nil
	}

	err = repo.QuestionRepo.RecoverQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}

	// update user's question count
	userQuestionCount, err := QuestionCommonServicer.GetUserQuestionCount(ctx, questionInfo.UserID)
	if err != nil {
		glog.Slog.Error("user GetUserQuestionCount error", err.Error())
	} else {
		err = UserCommonServicer.UpdateQuestionCount(ctx, questionInfo.UserID, userQuestionCount)
		if err != nil {
			glog.Slog.Error("user IncreaseQuestionCount error", err.Error())
		}
	}

	// update tag's question count
	if err = TagServicer.RemoveTagRelListByObjectID(ctx, questionInfo.ID); err != nil {
		glog.Slog.Errorf("remove tag rel list by object id error %v", err)
	}

	tagIDs := make([]string, 0)
	tags, err := TagServicer.GetObjectEntityTag(ctx, questionInfo.ID)
	if err != nil {
		return err
	}
	for _, v := range tags {
		tagIDs = append(tagIDs, v.ID)
	}
	if len(tagIDs) > 0 {
		if err = TagServicer.RefreshTagQuestionCount(ctx, tagIDs); err != nil {
			glog.Slog.Errorf("update tag's question count failed, %v", err)
		}
	}

	ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		TriggerUserID:    converter.StringToInt64(req.UserID),
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionUndeleted,
	})
	return nil
}

func (qs *QuestionService) UpdateQuestionInviteUser(ctx context.Context, req *schema.QuestionUpdateInviteUser) (err error) {
	originQuestion, exist, err := repo.QuestionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}

	//verify invite user
	inviteUserInfoList, err := UserCommonServicer.BatchGetUserBasicInfoByUserNames(ctx, req.InviteUser)
	if err != nil {
		glog.Slog.Error("BatchGetUserBasicInfoByUserNames error", err.Error())
	}
	inviteUserIDs := make([]string, 0)
	for _, item := range req.InviteUser {
		_, ok := inviteUserInfoList[item]
		if ok {
			inviteUserIDs = append(inviteUserIDs, inviteUserInfoList[item].ID)
		}
	}
	inviteUserStr := ""
	inviteUserByte, err := json.Marshal(inviteUserIDs)
	if err != nil {
		glog.Slog.Error("json.Marshal error", err.Error())
		inviteUserStr = "[]"
	} else {
		inviteUserStr = string(inviteUserByte)
	}
	question := &entity.Question{}
	question.ID = uid.DeShortID(req.ID)
	question.InviteUserID = inviteUserStr

	saveerr := repo.QuestionRepo.UpdateQuestion(ctx, question, []string{"invite_user_id"})
	if saveerr != nil {
		return saveerr
	}
	//send notification
	oldInviteUserIDsStr := originQuestion.InviteUserID
	oldInviteUserIDs := make([]string, 0)
	needSendNotificationUserIDs := make([]string, 0)
	if oldInviteUserIDsStr != "" {
		err = json.Unmarshal([]byte(oldInviteUserIDsStr), &oldInviteUserIDs)
		if err == nil {
			needSendNotificationUserIDs = converter.ArrayNotInArray(oldInviteUserIDs, inviteUserIDs)
		}
	} else {
		needSendNotificationUserIDs = inviteUserIDs
	}
	go qs.notificationInviteUser(ctx, needSendNotificationUserIDs, originQuestion.ID, originQuestion.Title, req.UserID)

	return nil
}

func (qs *QuestionService) notificationInviteUser(
	ctx context.Context, invitedUserIDs []string, questionID, questionTitle, questionUserID string) {
	inviter, exist, err := UserCommonServicer.GetUserBasicInfoByID(ctx, questionUserID)
	if err != nil {
		glog.Slog.Error(err)
		return
	}
	if !exist {
		log.Warnf("user %s not found", questionUserID)
		return
	}

	users, err := repo.UserRepo.BatchGetByID(ctx, invitedUserIDs)
	if err != nil {
		log.Error(err)
		return
	}
	invitee := make(map[string]*entity.User, len(users))
	for _, user := range users {
		invitee[user.ID] = user
	}
	for _, userID := range invitedUserIDs {
		msg := &schema.NotificationMsg{
			ReceiverUserID: userID,
			TriggerUserID:  questionUserID,
			Type:           schema.NotificationTypeInbox,
			ObjectID:       questionID,
		}
		msg.ObjectType = constant.QuestionObjectType
		msg.NotificationAction = constant.NotificationInvitedYouToAnswer
		NotificationQueueService.Send(ctx, msg)

		receiverUserInfo, ok := invitee[userID]
		if !ok {
			log.Warnf("user %s not found", userID)
			return
		}
		externalNotificationMsg := &schema.ExternalNotificationMsg{
			ReceiverUserID: receiverUserInfo.ID,
			ReceiverEmail:  receiverUserInfo.EMail,
			ReceiverLang:   receiverUserInfo.Language,
		}
		rawData := &schema.NewInviteAnswerTemplateRawData{
			InviterDisplayName: inviter.DisplayName,
			QuestionTitle:      questionTitle,
			QuestionID:         questionID,
			UnsubscribeCode:    token.GenerateToken(),
		}
		externalNotificationMsg.NewInviteAnswerTemplateRawData = rawData
		ExternalNotificationQueueService.Send(ctx, externalNotificationMsg)
	}
}

// UpdateQuestion update question
func (qs *QuestionService) UpdateQuestion(ctx context.Context, req *schema.QuestionUpdate) (questionInfo any, err error) {
	var canUpdate bool
	questionInfo = &schema.QuestionInfo{}

	_, existUnreviewed, err := RevisionComServicer.ExistUnreviewedByObjectID(ctx, req.ID)
	if err != nil {
		return
	}
	if existUnreviewed {
		err = errors.BadRequest(reason.QuestionCannotUpdate)
		return
	}

	dbinfo, has, err := repo.QuestionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return
	}
	if !has {
		return
	}
	if dbinfo.Status == entity.QuestionStatusDeleted {
		err = errors.BadRequest(reason.QuestionCannotUpdate)
		return nil, err
	}

	now := time.Now()
	question := &entity.Question{}
	question.Title = req.Title
	question.OriginalText = req.Content
	question.ParsedText = req.HTML
	question.ID = uid.DeShortID(req.ID)
	question.UpdatedAt = now
	question.PostUpdateTime = now
	question.UserID = dbinfo.UserID
	question.LastEditUserID = req.UserID

	oldTags, tagerr := TagServicer.GetObjectEntityTag(ctx, question.ID)
	if tagerr != nil {
		return questionInfo, tagerr
	}

	tagNameList := make([]string, 0)
	oldtagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tag.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		tagNameList = append(tagNameList, tag.SlugName)
	}
	for _, tag := range oldTags {
		oldtagNameList = append(oldtagNameList, tag.SlugName)
	}

	isChange := TagServicer.CheckTagsIsChange(ctx, tagNameList, oldtagNameList)

	//If the content is the same, ignore it
	if dbinfo.Title == req.Title && dbinfo.OriginalText == req.Content && !isChange {
		return
	}

	Tags, tagerr := TagServicer.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return questionInfo, tagerr
	}

	// if user can not use reserved tag, old reserved tag can not be removed and new reserved tag can not be added.
	if !req.CanUseReservedTag {
		CheckOldTag, CheckNewTag, CheckOldTaglist, CheckNewTaglist := qs.CheckChangeReservedTag(ctx, oldTags, Tags)
		if !CheckOldTag {
			errMsg := fmt.Sprintf(`The reserved tag "%s" must be present.`,
				strings.Join(CheckOldTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
		if !CheckNewTag {
			errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
				strings.Join(CheckNewTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
	}
	// Check whether mandatory labels are selected
	recommendExist, err := TagServicer.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(utils.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	//Administrators and themselves do not need to be audited

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   question.UserID,
		ObjectID: question.ID,
		Title:    question.Title,
		Log:      req.EditSummary,
	}

	if req.NoNeedReview {
		canUpdate = true
	}

	// It's not you or the administrator that needs to be reviewed
	if !canUpdate {
		revisionDTO.Status = entity.RevisionUnreviewedStatus
		revisionDTO.UserID = req.UserID //use revision userid
	} else {
		//Direct modification
		revisionDTO.Status = entity.RevisionReviewPassStatus
		//update question to db
		saveerr := repo.QuestionRepo.UpdateQuestion(ctx, question, []string{"title", "original_text", "parsed_text", "updated_at", "post_update_time", "last_edit_user_id"})
		if saveerr != nil {
			return questionInfo, saveerr
		}
		objectTagData := schema.TagChange{}
		objectTagData.ObjectID = question.ID
		objectTagData.Tags = req.Tags
		objectTagData.UserID = req.UserID
		tagerr := qs.ChangeTag(ctx, &objectTagData)
		if err != nil {
			return questionInfo, tagerr
		}
	}

	questionWithTagsRevision, err := qs.changeQuestionToRevision(ctx, question, Tags)
	if err != nil {
		return nil, err
	}
	infoJSON, _ := json.Marshal(questionWithTagsRevision)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := RevisionComServicer.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return
	}
	if canUpdate {
		ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         question.ID,
			ActivityTypeKey:  constant.ActQuestionEdited,
			RevisionID:       revisionID,
			OriginalObjectID: question.ID,
		})
	}

	questionInfo, err = qs.GetQuestion(ctx, question.ID, question.UserID, req.QuestionPermission)
	return
}

// GetQuestion get question one
func (qs *QuestionService) GetQuestion(ctx context.Context, questionID, userID string,
	per schema.QuestionPermission) (resp *schema.QuestionInfo, err error) {

	//
	question, err := QuestionCommonServicer.Info(ctx, questionID, userID)
	if err != nil {
		return
	}
	// If the question is deleted, only the administrator and the author can view it
	if question.Status == entity.QuestionStatusDeleted && !per.CanReopen && question.UserID != userID {
		return nil, errors.NotFound(reason.QuestionNotFound)
	}
	//问题没关闭
	if question.Status != entity.QuestionStatusClosed {
		//不需要重新开启
		per.CanReopen = false
	}
	//问题关闭
	if question.Status == entity.QuestionStatusClosed {
		//不需要关闭
		per.CanClose = false
	}
	//todo pin啥意思
	if question.Pin == entity.QuestionPin {
		per.CanPin = false
		per.CanHide = false
	}
	if question.Pin == entity.QuestionUnPin {
		per.CanUnPin = false
	}
	if question.Show == entity.QuestionShow {
		per.CanShow = false
	}
	if question.Show == entity.QuestionHide {
		per.CanHide = false
		per.CanPin = false
	}

	if question.Status == entity.QuestionStatusDeleted {
		operation := &schema.Operation{}
		operation.Msg = translator.Tr(utils.GetLangByCtx(ctx), reason.QuestionAlreadyDeleted)
		operation.Level = schema.OperationLevelDanger
		question.Operation = operation
	}

	question.Description = htmltext.FetchExcerpt(question.HTML, "...", 240)
	question.MemberActions = permission.GetQuestionPermission(ctx, userID, question.UserID, question.Status,
		per.CanEdit, per.CanDelete,
		per.CanClose, per.CanReopen, per.CanPin, per.CanHide, per.CanUnPin, per.CanShow,
		per.CanRecover)
	question.ExtendsActions = permission.GetQuestionExtendsPermission(ctx, per.CanInviteOtherToAnswer)
	return question, nil
}

// GetQuestionAndAddPV get question one
func (qs *QuestionService) GetQuestionAndAddPV(ctx context.Context, questionID, loginUserID string,
	per schema.QuestionPermission) (resp *schema.QuestionInfo, err error) {

	err = QuestionCommonServicer.UpdatePv(ctx, questionID)
	if err != nil {
		log.Error(err)
	}
	return qs.GetQuestion(ctx, questionID, loginUserID, per)
}

func (qs *QuestionService) InviteUserInfo(ctx context.Context, questionID string) (inviteList []*schema.UserBasicInfo, err error) {
	return QuestionCommonServicer.InviteUserInfo(ctx, questionID)
}

func (qs *QuestionService) ChangeTag(ctx context.Context, objectTagData *schema.TagChange) error {
	return TagServicer.ObjectChangeTag(ctx, objectTagData)
}

func (qs *QuestionService) CheckChangeReservedTag(ctx context.Context, oldobjectTagData, objectTagData []*entity.Tag) (bool, bool, []string, []string) {
	return TagServicer.CheckChangeReservedTag(ctx, oldobjectTagData, objectTagData)
}

// PersonalQuestionPage get question list by user
func (qs *QuestionService) PersonalQuestionPage(ctx context.Context, req *schema.PersonalQuestionPageReq) (
	pageModel *pager.PageModel, err error) {

	userinfo, exist, err := UserCommonServicer.GetUserBasicInfoByUserName(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = req.OrderCond
	search.Page = req.Page
	search.PageSize = req.PageSize
	search.UserIDBeSearched = userinfo.ID
	search.LoginUserID = req.LoginUserID
	questionList, total, err := qs.GetQuestionPage(ctx, search)
	if err != nil {
		return nil, err
	}
	userQuestionInfoList := make([]*schema.UserQuestionInfo, 0)
	for _, item := range questionList {
		info := &schema.UserQuestionInfo{}
		_ = copier.Copy(info, item)
		status, ok := entity.AdminQuestionSearchStatusIntToString[item.Status]
		if ok {
			info.Status = status
		}
		userQuestionInfoList = append(userQuestionInfoList, info)
	}
	return pager.NewPageModel(total, userQuestionInfoList), nil
}

func (qs *QuestionService) PersonalAnswerPage(ctx context.Context, req *schema.PersonalAnswerPageReq) (
	pageModel *pager.PageModel, err error) {
	userinfo, exist, err := UserCommonServicer.GetUserBasicInfoByUserName(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	answersearch := &entity.AnswerSearch{}
	answersearch.UserID = userinfo.ID
	answersearch.PageSize = req.PageSize
	answersearch.Page = req.Page
	if req.OrderCond == "newest" {
		answersearch.Order = entity.AnswerSearchOrderByTime
	} else {
		answersearch.Order = entity.AnswerSearchOrderByDefault
	}
	questionIDs := make([]string, 0)
	answerList, total, err := AnswerCommonServicer.Search(ctx, answersearch)
	if err != nil {
		return nil, err
	}

	answerlist := make([]*schema.AnswerInfo, 0)
	userAnswerlist := make([]*schema.UserAnswerInfo, 0)
	for _, item := range answerList {
		answerinfo := AnswerCommonServicer.ShowFormat(ctx, item)
		answerlist = append(answerlist, answerinfo)
		questionIDs = append(questionIDs, uid.DeShortID(item.QuestionID))
	}
	questionMaps, err := QuestionCommonServicer.FindInfoByID(ctx, questionIDs, req.LoginUserID)
	if err != nil {
		return nil, err
	}

	for _, item := range answerlist {
		_, ok := questionMaps[item.QuestionID]
		if ok {
			item.QuestionInfo = questionMaps[item.QuestionID]
		} else {
			continue
		}
		info := &schema.UserAnswerInfo{}
		_ = copier.Copy(info, item)
		info.AnswerID = item.ID
		info.QuestionID = item.QuestionID
		if item.QuestionInfo.Status == entity.QuestionStatusDeleted {
			info.QuestionInfo.Title = "Deleted question"

		}
		userAnswerlist = append(userAnswerlist, info)
	}

	return pager.NewPageModel(total, userAnswerlist), nil
}

// PersonalCollectionPage get collection list by user
func (qs *QuestionService) PersonalCollectionPage(ctx context.Context, req *schema.PersonalCollectionPageReq) (
	pageModel *pager.PageModel, err error) {
	list := make([]*schema.QuestionInfo, 0)
	collectionSearch := &entity.CollectionSearch{}
	collectionSearch.UserID = req.UserID
	collectionSearch.Page = req.Page
	collectionSearch.PageSize = req.PageSize
	collectionList, total, err := CollectionCommon.SearchList(ctx, collectionSearch)
	if err != nil {
		return nil, err
	}
	questionIDs := make([]string, 0)
	for _, item := range collectionList {
		questionIDs = append(questionIDs, item.ObjectID)
	}

	questionMaps, err := QuestionCommonServicer.FindInfoByID(ctx, questionIDs, req.UserID)
	if err != nil {
		return nil, err
	}
	for _, id := range questionIDs {
		if utils.GetEnableShortID(ctx) {
			id = uid.EnShortID(id)
		}
		_, ok := questionMaps[id]
		if ok {
			questionMaps[id].LastAnsweredUserInfo = nil
			questionMaps[id].UpdateUserInfo = nil
			questionMaps[id].Content = ""
			questionMaps[id].HTML = ""
			if questionMaps[id].Status == entity.QuestionStatusDeleted {
				questionMaps[id].Title = "Deleted question"
			}
			list = append(list, questionMaps[id])
		}
	}

	return pager.NewPageModel(total, list), nil
}

func (qs *QuestionService) SearchUserTopList(ctx context.Context, userName string, loginUserID string) ([]*schema.UserQuestionInfo, []*schema.UserAnswerInfo, error) {
	answerlist := make([]*schema.AnswerInfo, 0)

	userAnswerlist := make([]*schema.UserAnswerInfo, 0)
	userQuestionlist := make([]*schema.UserQuestionInfo, 0)

	userinfo, Exist, err := UserCommonServicer.GetUserBasicInfoByUserName(ctx, userName)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	if !Exist {
		return userQuestionlist, userAnswerlist, nil
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = "score"
	search.Page = 0
	search.PageSize = 5
	search.UserIDBeSearched = userinfo.ID
	search.LoginUserID = loginUserID
	questionlist, _, err := qs.GetQuestionPage(ctx, search)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	answersearch := &entity.AnswerSearch{}
	answersearch.UserID = userinfo.ID
	answersearch.PageSize = 5
	answersearch.Order = entity.AnswerSearchOrderByVote
	questionIDs := make([]string, 0)
	answerList, _, err := AnswerCommonServicer.Search(ctx, answersearch)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	for _, item := range answerList {
		answerinfo := AnswerCommonServicer.ShowFormat(ctx, item)
		answerlist = append(answerlist, answerinfo)
		questionIDs = append(questionIDs, item.QuestionID)
	}
	questionMaps, err := QuestionCommonServicer.FindInfoByID(ctx, questionIDs, loginUserID)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	for _, item := range answerlist {
		_, ok := questionMaps[item.QuestionID]
		if ok {
			item.QuestionInfo = questionMaps[item.QuestionID]
		}
	}

	for _, item := range questionlist {
		info := &schema.UserQuestionInfo{}
		_ = copier.Copy(info, item)
		info.UrlTitle = htmltext.UrlTitle(info.Title)
		userQuestionlist = append(userQuestionlist, info)
	}

	for _, item := range answerlist {
		info := &schema.UserAnswerInfo{}
		_ = copier.Copy(info, item)
		info.AnswerID = item.ID
		info.QuestionID = item.QuestionID
		info.QuestionInfo.UrlTitle = htmltext.UrlTitle(info.QuestionInfo.Title)
		userAnswerlist = append(userAnswerlist, info)
	}

	return userQuestionlist, userAnswerlist, nil
}

// GetQuestionsByTitle get questions by title
func (qs *QuestionService) GetQuestionsByTitle(ctx context.Context, title string) (
	resp []*schema.QuestionBaseInfo, err error) {
	resp = make([]*schema.QuestionBaseInfo, 0)
	if len(title) == 0 {
		return resp, nil
	}
	questions, err := repo.QuestionRepo.GetQuestionsByTitle(ctx, title, 10)
	if err != nil {
		return resp, err
	}
	for _, question := range questions {
		item := &schema.QuestionBaseInfo{}
		item.ID = question.ID
		item.Title = question.Title
		item.UrlTitle = htmltext.UrlTitle(question.Title)
		item.ViewCount = question.ViewCount
		item.AnswerCount = question.AnswerCount
		item.CollectionCount = question.CollectionCount
		item.FollowCount = question.FollowCount
		status, ok := entity.AdminQuestionSearchStatusIntToString[question.Status]
		if ok {
			item.Status = status
		}
		if question.AcceptedAnswerID != "0" {
			item.AcceptedAnswer = true
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// SimilarQuestion
func (qs *QuestionService) SimilarQuestion(ctx context.Context, questionID string, loginUserID string) ([]*schema.QuestionPageResp, int64, error) {
	question, err := QuestionCommonServicer.Info(ctx, questionID, loginUserID)
	if err != nil {
		return nil, 0, nil
	}
	tagNames := make([]string, 0, len(question.Tags))
	for _, tag := range question.Tags {
		tagNames = append(tagNames, tag.SlugName)
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = "frequent"
	search.Page = 0
	search.PageSize = 6
	if len(tagNames) > 0 {
		search.Tag = tagNames[0]
	}
	search.LoginUserID = loginUserID
	return qs.GetQuestionPage(ctx, search)
}

// GetQuestionPage query questions page
func (qs *QuestionService) GetQuestionPage(ctx context.Context, req *schema.QuestionPageReq) (
	questions []*schema.QuestionPageResp, total int64, err error) {
	questions = make([]*schema.QuestionPageResp, 0)

	// query by tag condition
	if len(req.Tag) > 0 {
		tagInfo, exist, err := TagServicer.GetTagBySlugName(ctx, strings.ToLower(req.Tag))
		if err != nil {
			return nil, 0, err
		}
		if exist {
			req.TagID = tagInfo.ID
		}
	}

	// query by user condition
	if req.Username != "" {
		userinfo, exist, err := UserCommonServicer.GetUserBasicInfoByUserName(ctx, req.Username)
		if err != nil {
			return nil, 0, err
		}
		if !exist {
			return questions, 0, nil
		}
		req.UserIDBeSearched = userinfo.ID
	}

	questionList, total, err := repo.QuestionRepo.GetQuestionPage(ctx, req.Page, req.PageSize,
		req.UserIDBeSearched, req.TagID, req.OrderCond, req.InDays)
	if err != nil {
		return nil, 0, err
	}
	questions, err = QuestionCommonServicer.FormatQuestionsPage(ctx, questionList, req.LoginUserID, req.OrderCond)
	if err != nil {
		return nil, 0, err
	}
	return questions, total, nil
}

func (qs *QuestionService) AdminSetQuestionStatus(ctx context.Context, req *schema.AdminUpdateQuestionStatusReq) error {
	setStatus, ok := entity.AdminQuestionSearchStatus[req.Status]
	if !ok {
		return errors.BadRequest(reason.RequestFormatError)
	}
	questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}
	err = repo.QuestionRepo.UpdateQuestionStatus(ctx, questionInfo.ID, setStatus)
	if err != nil {
		return err
	}

	msg := &schema.NotificationMsg{}
	if setStatus == entity.QuestionStatusDeleted {
		// #2372 In order to simplify the process and complexity, as well as to consider if it is in-house,
		// facing the problem of recovery.
		//err = qs.answerActivityService.DeleteQuestion(ctx, questionInfo.ID, questionInfo.CreatedAt, questionInfo.VoteCount)
		//if err != nil {
		//	log.Errorf("admin delete question then rank rollback error %s", err.Error())
		//}
		ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionDeleted,
		})
		msg.NotificationAction = constant.NotificationYourQuestionWasDeleted
	}
	if setStatus == entity.QuestionStatusAvailable && questionInfo.Status == entity.QuestionStatusClosed {
		ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionReopened,
		})
	}
	if setStatus == entity.QuestionStatusClosed && questionInfo.Status != entity.QuestionStatusClosed {
		ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionClosed,
		})
		msg.NotificationAction = constant.NotificationYourQuestionIsClosed
	}
	// recover
	if setStatus == entity.QuestionStatusAvailable && questionInfo.Status == entity.QuestionStatusDeleted {
		ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionUndeleted,
		})
	}

	if len(msg.NotificationAction) > 0 {
		msg.ObjectID = questionInfo.ID
		msg.Type = schema.NotificationTypeInbox
		msg.ReceiverUserID = questionInfo.UserID
		msg.TriggerUserID = req.UserID
		msg.ObjectType = constant.QuestionObjectType
		NotificationQueueService.Send(ctx, msg)
	}
	return nil
}

func (qs *QuestionService) AdminQuestionPage(
	ctx context.Context, req *schema.AdminQuestionPageReq) (
	resp *pager.PageModel, err error) {

	list := make([]*schema.AdminQuestionInfo, 0)
	questionList, count, err := repo.QuestionRepo.AdminQuestionPage(ctx, req)
	if err != nil {
		return nil, err
	}

	userIds := make([]string, 0)
	for _, info := range questionList {
		item := &schema.AdminQuestionInfo{}
		_ = copier.Copy(item, info)
		item.CreateTime = info.CreatedAt.Unix()
		item.UpdateTime = info.PostUpdateTime.Unix()
		item.EditTime = info.UpdatedAt.Unix()
		list = append(list, item)
		userIds = append(userIds, info.UserID)
	}
	userInfoMap, err := UserCommonServicer.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		if u, ok := userInfoMap[item.UserID]; ok {
			item.UserInfo = u
		}
	}
	return pager.NewPageModel(count, list), nil
}

// AdminAnswerPage search answer list
func (qs *QuestionService) AdminAnswerPage(ctx context.Context, req *schema.AdminAnswerPageReq) (
	resp *pager.PageModel, err error) {
	answerList, count, err := AnswerCommonServicer.AdminSearchList(ctx, req)
	if err != nil {
		return nil, err
	}

	questionIDs := make([]string, 0)
	userIds := make([]string, 0)
	answerResp := make([]*schema.AdminAnswerInfo, 0)
	for _, item := range answerList {
		answerInfo := AnswerCommonServicer.AdminShowFormat(ctx, item)
		answerResp = append(answerResp, answerInfo)
		questionIDs = append(questionIDs, item.QuestionID)
		userIds = append(userIds, item.UserID)
	}
	userInfoMap, err := UserCommonServicer.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return nil, err
	}
	questionMaps, err := QuestionCommonServicer.FindInfoByID(ctx, questionIDs, req.LoginUserID)
	if err != nil {
		return nil, err
	}

	for _, item := range answerResp {
		if q, ok := questionMaps[item.QuestionID]; ok {
			item.QuestionInfo.Title = q.Title
		}
		if u, ok := userInfoMap[item.UserID]; ok {
			item.UserInfo = u
		}
	}
	return pager.NewPageModel(count, answerResp), nil
}

func (qs *QuestionService) changeQuestionToRevision(ctx context.Context, questionInfo *entity.Question, tags []*entity.Tag) (
	questionRevision *entity.QuestionWithTagsRevision, err error) {

	questionRevision = &entity.QuestionWithTagsRevision{}
	questionRevision.Question = *questionInfo

	for _, tag := range tags {
		item := &entity.TagSimpleInfoForRevision{}
		_ = copier.Copy(item, tag)
		questionRevision.Tags = append(questionRevision.Tags, item)
	}
	return questionRevision, nil
}

func (qs *QuestionService) SitemapCron(ctx context.Context) {
	siteSeo := site.Config.GetSiteSeo()
	ctx = context.WithValue(ctx, constant.ShortIDFlag, siteSeo.IsShortLink())
	QuestionCommonServicer.SitemapCron(ctx)
}
