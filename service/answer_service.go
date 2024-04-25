package service

import (
	"context"
	"encoding/json"
	constant2 "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity "github.com/lawyer/commons/entity"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/repo"
	"github.com/lawyer/service/permission"
	role2 "github.com/lawyer/service/role"
	"time"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/token"
	"github.com/lawyer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// AnswerService user service
type AnswerService struct {
}

func NewAnswerService() *AnswerService {
	return &AnswerService{}
}

// RemoveAnswer delete answer
func (as *AnswerService) RemoveAnswer(ctx context.Context, req *schema.RemoveAnswerReq) (err error) {
	answerInfo, exist, err := repo.AnswerRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	// if the status is deleted, return directly
	if answerInfo.Status == entity.AnswerStatusDeleted {
		return nil
	}
	roleID, err := services.UserRoleRelService.GetUserRole(ctx, req.UserID)
	if err != nil {
		return err
	}
	if roleID != role2.RoleAdminID && roleID != role2.RoleModeratorID {
		if answerInfo.UserID != req.UserID {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}
		if answerInfo.VoteCount > 0 {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}
		if answerInfo.Accepted == schema.AnswerAcceptedEnable {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}
		_, exist, err := repo.QuestionRepo.GetQuestion(ctx, answerInfo.QuestionID)
		if err != nil {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}
		if !exist {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}

	}

	err = repo.AnswerRepo.RemoveAnswer(ctx, req.ID)
	if err != nil {
		return err
	}

	// user add question count
	err = services.QuestionCommon.UpdateAnswerCount(ctx, answerInfo.QuestionID)
	if err != nil {
		log.Error("IncreaseAnswerCount error", err.Error())
	}
	userAnswerCount, err := repo.AnswerRepo.GetCountByUserID(ctx, answerInfo.UserID)
	if err != nil {
		log.Error("GetCountByUserID error", err.Error())
	}
	err = services.UserCommon.UpdateAnswerCount(ctx, answerInfo.UserID, int(userAnswerCount))
	if err != nil {
		log.Error("user IncreaseAnswerCount error", err.Error())
	}
	// #2372 In order to simplify the process and complexity, as well as to consider if it is in-house,
	// facing the problem of recovery.
	//err = initServer.AnswerActivityService.DeleteAnswer(ctx, answerInfo.ID, answerInfo.CreatedAt, answerInfo.VoteCount)
	//if err != nil {
	//	log.Errorf("delete answer activity change failed: %s", err.Error())
	//}
	services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         answerInfo.ID,
		OriginalObjectID: answerInfo.ID,
		ActivityTypeKey:  constant2.ActAnswerDeleted,
	})
	return
}

// RecoverAnswer recover deleted answer
func (as *AnswerService) RecoverAnswer(ctx context.Context, req *schema.RecoverAnswerReq) (err error) {
	answerInfo, exist, err := repo.AnswerRepo.GetByID(ctx, req.AnswerID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.AnswerNotFound)
	}
	if answerInfo.Status != entity.AnswerStatusDeleted {
		return nil
	}
	if err = repo.AnswerRepo.RecoverAnswer(ctx, req.AnswerID); err != nil {
		return err
	}

	if err = services.QuestionCommon.UpdateAnswerCount(ctx, answerInfo.QuestionID); err != nil {
		log.Errorf("update answer count failed: %s", err.Error())
	}
	userAnswerCount, err := repo.AnswerRepo.GetCountByUserID(ctx, answerInfo.UserID)
	if err != nil {
		log.Errorf("get user answer count failed: %s", err.Error())
	} else {
		err = services.UserCommon.UpdateAnswerCount(ctx, answerInfo.UserID, int(userAnswerCount))
		if err != nil {
			log.Errorf("update user answer count failed: %s", err.Error())
		}
	}
	services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		TriggerUserID:    converter.StringToInt64(req.UserID),
		ObjectID:         answerInfo.ID,
		OriginalObjectID: answerInfo.ID,
		ActivityTypeKey:  constant2.ActAnswerUndeleted,
	})
	return nil
}

func (as *AnswerService) Insert(ctx context.Context, req *schema.AnswerAddReq) (string, error) {
	questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.BadRequest(reason.QuestionNotFound)
	}
	if questionInfo.Status == entity.QuestionStatusClosed || questionInfo.Status == entity.QuestionStatusDeleted {
		err = errors.BadRequest(reason.AnswerCannotAddByClosedQuestion)
		return "", err
	}
	insertData := new(entity.Answer)
	insertData.UserID = req.UserID
	insertData.OriginalText = req.Content
	insertData.ParsedText = req.HTML
	insertData.Accepted = schema.AnswerAcceptedFailed
	insertData.QuestionID = req.QuestionID
	insertData.RevisionID = "0"
	insertData.LastEditUserID = "0"
	insertData.Status = entity.AnswerStatusAvailable
	//insertData.UpdatedAt = now
	if err = repo.AnswerRepo.AddAnswer(ctx, insertData); err != nil {
		return "", err
	}
	err = services.QuestionCommon.UpdateAnswerCount(ctx, req.QuestionID)
	if err != nil {
		log.Error("IncreaseAnswerCount error", err.Error())
	}
	err = services.QuestionCommon.UpdateLastAnswer(ctx, req.QuestionID, uid.DeShortID(insertData.ID))
	if err != nil {
		log.Error("UpdateLastAnswer error", err.Error())
	}
	err = services.QuestionCommon.UpdatePostTime(ctx, req.QuestionID)
	if err != nil {
		return insertData.ID, err
	}
	userAnswerCount, err := repo.AnswerRepo.GetCountByUserID(ctx, req.UserID)
	if err != nil {
		log.Error("GetCountByUserID error", err.Error())
	}
	err = services.UserCommon.UpdateAnswerCount(ctx, req.UserID, int(userAnswerCount))
	if err != nil {
		log.Error("user IncreaseAnswerCount error", err.Error())
	}

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   insertData.UserID,
		ObjectID: insertData.ID,
		Title:    "",
	}
	infoJSON, _ := json.Marshal(insertData)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := services.RevisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return insertData.ID, err
	}
	as.notificationAnswerTheQuestion(ctx, questionInfo.UserID, questionInfo.ID, insertData.ID, req.UserID, questionInfo.Title,
		insertData.OriginalText)

	services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           insertData.UserID,
		ObjectID:         insertData.ID,
		OriginalObjectID: insertData.ID,
		ActivityTypeKey:  constant2.ActAnswerAnswered,
		RevisionID:       revisionID,
	})
	services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           insertData.UserID,
		ObjectID:         insertData.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant2.ActQuestionAnswered,
	})
	return insertData.ID, nil
}

func (as *AnswerService) Update(ctx context.Context, req *schema.AnswerUpdateReq) (string, error) {
	var canUpdate bool
	_, existUnreviewed, err := services.RevisionService.ExistUnreviewedByObjectID(ctx, req.ID)
	if err != nil {
		return "", err
	}
	if existUnreviewed {
		return "", errors.BadRequest(reason.AnswerCannotUpdate)
	}

	questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.BadRequest(reason.QuestionNotFound)
	}

	answerInfo, exist, err := repo.AnswerRepo.GetByID(ctx, req.ID)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.BadRequest(reason.AnswerNotFound)
	}

	if answerInfo.Status == entity.AnswerStatusDeleted {
		return "", errors.BadRequest(reason.AnswerCannotUpdate)
	}

	//If the content is the same, ignore it
	if answerInfo.OriginalText == req.Content {
		return "", nil
	}

	insertData := &entity.Answer{}
	insertData.ID = req.ID
	insertData.UserID = answerInfo.UserID
	insertData.QuestionID = req.QuestionID
	insertData.OriginalText = req.Content
	insertData.ParsedText = req.HTML
	insertData.UpdatedAt = time.Now()
	insertData.LastEditUserID = "0"
	if answerInfo.UserID != req.UserID {
		insertData.LastEditUserID = req.UserID
	}

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   req.UserID,
		ObjectID: req.ID,
		Log:      req.EditSummary,
	}

	if req.NoNeedReview || answerInfo.UserID == req.UserID {
		canUpdate = true
	}

	if !canUpdate {
		revisionDTO.Status = entity.RevisionUnreviewedStatus
	} else {
		if err = repo.AnswerRepo.UpdateAnswer(ctx, insertData, []string{"original_text", "parsed_text", "updated_at", "last_edit_user_id"}); err != nil {
			return "", err
		}
		err = services.QuestionCommon.UpdatePostTime(ctx, req.QuestionID)
		if err != nil {
			return insertData.ID, err
		}
		as.notificationUpdateAnswer(ctx, questionInfo.UserID, insertData.ID, req.UserID)
		revisionDTO.Status = entity.RevisionReviewPassStatus
	}

	infoJSON, _ := json.Marshal(insertData)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := services.RevisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return insertData.ID, err
	}
	if canUpdate {
		services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         insertData.ID,
			OriginalObjectID: insertData.ID,
			ActivityTypeKey:  constant2.ActAnswerEdited,
			RevisionID:       revisionID,
		})
	}

	return insertData.ID, nil
}

// AcceptAnswer accept answer
func (as *AnswerService) AcceptAnswer(ctx context.Context, req *schema.AcceptAnswerReq) (err error) {
	// find question
	questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}
	questionInfo.ID = uid.DeShortID(questionInfo.ID)
	if questionInfo.AcceptedAnswerID == req.AnswerID {
		return nil
	}

	// find answer
	var acceptedAnswerInfo *entity.Answer
	if len(req.AnswerID) > 1 {
		acceptedAnswerInfo, exist, err = repo.AnswerRepo.GetByID(ctx, req.AnswerID)
		if err != nil {
			return err
		}
		if !exist {
			return errors.BadRequest(reason.AnswerNotFound)
		}
		acceptedAnswerInfo.ID = uid.DeShortID(acceptedAnswerInfo.ID)
	}

	// update answers status
	if err = repo.AnswerRepo.UpdateAcceptedStatus(ctx, req.AnswerID, req.QuestionID); err != nil {
		return err
	}

	// update question status
	err = services.QuestionCommon.UpdateAccepted(ctx, req.QuestionID, req.AnswerID)
	if err != nil {
		log.Error("UpdateLastAnswer error", err.Error())
	}

	var oldAnswerInfo *entity.Answer
	if len(questionInfo.AcceptedAnswerID) > 1 {
		oldAnswerInfo, _, err = repo.AnswerRepo.GetByID(ctx, questionInfo.AcceptedAnswerID)
		if err != nil {
			return err
		}
		oldAnswerInfo.ID = uid.DeShortID(oldAnswerInfo.ID)
	}

	as.updateAnswerRank(ctx, req.UserID, questionInfo, acceptedAnswerInfo, oldAnswerInfo)
	return nil
}

func (as *AnswerService) updateAnswerRank(ctx context.Context, userID string,
	questionInfo *entity.Question, newAnswerInfo *entity.Answer, oldAnswerInfo *entity.Answer,
) {
	// if this question is already been answered, should cancel old answer rank
	if oldAnswerInfo != nil {
		err := services.AnswerActivityService.CancelAcceptAnswer(ctx, userID,
			questionInfo.AcceptedAnswerID, questionInfo.ID, questionInfo.UserID, oldAnswerInfo.UserID)
		if err != nil {
			log.Error(err)
		}
	}
	if newAnswerInfo != nil {
		err := services.AnswerActivityService.AcceptAnswer(ctx, userID, newAnswerInfo.ID,
			questionInfo.ID, questionInfo.UserID, newAnswerInfo.UserID, newAnswerInfo.UserID == questionInfo.UserID)
		if err != nil {
			log.Error(err)
		}
	}
}

func (as *AnswerService) Get(ctx context.Context, answerID, loginUserID string) (*schema.AnswerInfo, *schema.QuestionInfo, bool, error) {
	answerInfo, has, err := repo.AnswerRepo.GetByID(ctx, answerID)
	if err != nil {
		return nil, nil, has, err
	}
	info := as.ShowFormat(ctx, answerInfo)
	// todo questionFunc
	questionInfo, err := services.QuestionCommon.Info(ctx, answerInfo.QuestionID, loginUserID)
	if err != nil {
		return nil, nil, has, err
	}
	// todo UserFunc

	userIds := make([]string, 0)
	userIds = append(userIds, answerInfo.UserID)
	userIds = append(userIds, answerInfo.LastEditUserID)
	userInfoMap, err := services.UserCommon.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return nil, nil, has, err
	}

	_, ok := userInfoMap[answerInfo.UserID]
	if ok {
		info.UserInfo = userInfoMap[answerInfo.UserID]
	}
	_, ok = userInfoMap[answerInfo.LastEditUserID]
	if ok {
		info.UpdateUserInfo = userInfoMap[answerInfo.LastEditUserID]
	}

	if loginUserID == "" {
		return info, questionInfo, has, nil
	}

	info.VoteStatus = repo.VoteRepo.GetVoteStatus(ctx, answerID, loginUserID)

	collectedMap, err := services.CollectionCommon.SearchObjectCollected(ctx, loginUserID, []string{answerInfo.ID})
	if err != nil {
		return nil, nil, has, err
	}
	if len(collectedMap) > 0 {
		info.Collected = true
	}

	return info, questionInfo, has, nil
}

func (as *AnswerService) GetCountByUserIDQuestionID(ctx context.Context, userId string, questionId string) (ids []string, err error) {
	return repo.AnswerRepo.GetIDsByUserIDAndQuestionID(ctx, userId, questionId)
}

func (as *AnswerService) AdminSetAnswerStatus(ctx context.Context, req *schema.AdminUpdateAnswerStatusReq) error {
	setStatus, ok := entity.AdminAnswerSearchStatus[req.Status]
	if !ok {
		return errors.BadRequest(reason.RequestFormatError)
	}
	answerInfo, exist, err := repo.AnswerRepo.GetAnswer(ctx, req.AnswerID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.AnswerNotFound)
	}
	err = repo.AnswerRepo.UpdateAnswerStatus(ctx, answerInfo.ID, setStatus)
	if err != nil {
		return err
	}

	if setStatus == entity.AnswerStatusDeleted {
		// #2372 In order to simplify the process and complexity, as well as to consider if it is in-house,
		// facing the problem of recovery.
		//err = initServer.AnswerActivityService.DeleteAnswer(ctx, answerInfo.ID, answerInfo.CreatedAt, answerInfo.VoteCount)
		//if err != nil {
		//	log.Errorf("admin delete question then rank rollback error %s", err.Error())
		//}
		services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         answerInfo.ID,
			OriginalObjectID: answerInfo.ID,
			ActivityTypeKey:  constant2.ActAnswerDeleted,
		})

		msg := &schema.NotificationMsg{}
		msg.ObjectID = answerInfo.ID
		msg.Type = schema.NotificationTypeInbox
		msg.ReceiverUserID = answerInfo.UserID
		msg.TriggerUserID = answerInfo.UserID
		msg.ObjectType = constant2.AnswerObjectType
		msg.NotificationAction = constant2.NotificationYourAnswerWasDeleted
		services.NotificationQueueService.Send(ctx, msg)
	}

	// recover
	if setStatus == entity.QuestionStatusAvailable && answerInfo.Status == entity.QuestionStatusDeleted {
		services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         answerInfo.ID,
			OriginalObjectID: answerInfo.ID,
			ActivityTypeKey:  constant2.ActAnswerUndeleted,
		})
	}
	return nil
}

func (as *AnswerService) SearchList(ctx context.Context, req *schema.AnswerListReq) ([]*schema.AnswerInfo, int64, error) {
	list := make([]*schema.AnswerInfo, 0)
	dbSearch := entity.AnswerSearch{}
	dbSearch.QuestionID = req.QuestionID
	dbSearch.Page = req.Page
	dbSearch.PageSize = req.PageSize
	dbSearch.Order = req.Order
	dbSearch.IncludeDeleted = req.CanDelete
	dbSearch.LoginUserID = req.UserID
	answerOriginalList, count, err := repo.AnswerRepo.SearchList(ctx, &dbSearch)
	if err != nil {
		return list, count, err
	}
	answerList, err := as.SearchFormatInfo(ctx, answerOriginalList, req)
	if err != nil {
		return answerList, count, err
	}
	return answerList, count, nil
}

func (as *AnswerService) SearchFormatInfo(ctx context.Context, answers []*entity.Answer, req *schema.AnswerListReq) (
	[]*schema.AnswerInfo, error) {
	list := make([]*schema.AnswerInfo, 0)
	objectIDs := make([]string, 0)
	userIDs := make([]string, 0)
	for _, info := range answers {
		item := as.ShowFormat(ctx, info)
		list = append(list, item)
		objectIDs = append(objectIDs, info.ID)
		userIDs = append(userIDs, info.UserID, info.LastEditUserID)
	}

	userInfoMap, err := services.UserCommon.BatchUserBasicInfoByID(ctx, userIDs)
	if err != nil {
		return list, err
	}
	for _, item := range list {
		item.UserInfo = userInfoMap[item.UserID]
		item.UpdateUserInfo = userInfoMap[item.UpdateUserID]
	}
	if len(req.UserID) == 0 {
		return list, nil
	}

	collectedMap, err := services.CollectionCommon.SearchObjectCollected(ctx, req.UserID, objectIDs)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		item.VoteStatus = repo.VoteRepo.GetVoteStatus(ctx, item.ID, req.UserID)
		item.Collected = collectedMap[item.ID]
		item.MemberActions = permission.GetAnswerPermission(ctx,
			req.UserID,
			item.UserID,
			item.Status,
			req.CanEdit,
			req.CanDelete,
			req.CanRecover)
	}
	return list, nil
}

func (as *AnswerService) ShowFormat(ctx context.Context, data *entity.Answer) *schema.AnswerInfo {
	return services.AnswerCommon.ShowFormat(ctx, data)
}

func (as *AnswerService) notificationUpdateAnswer(ctx context.Context, questionUserID, answerID, answerUserID string) {
	msg := &schema.NotificationMsg{
		TriggerUserID:  answerUserID,
		ReceiverUserID: questionUserID,
		Type:           schema.NotificationTypeInbox,
		ObjectID:       answerID,
	}
	msg.ObjectType = constant2.AnswerObjectType
	msg.NotificationAction = constant2.NotificationUpdateAnswer
	services.NotificationQueueService.Send(ctx, msg)
}

func (as *AnswerService) notificationAnswerTheQuestion(ctx context.Context,
	questionUserID, questionID, answerID, answerUserID, questionTitle, answerSummary string) {
	// If the question is answered by me, there is no notification for myself.
	if questionUserID == answerUserID {
		return
	}
	msg := &schema.NotificationMsg{
		TriggerUserID:  answerUserID,
		ReceiverUserID: questionUserID,
		Type:           schema.NotificationTypeInbox,
		ObjectID:       answerID,
	}
	msg.ObjectType = constant2.AnswerObjectType
	msg.NotificationAction = constant2.NotificationAnswerTheQuestion
	services.NotificationQueueService.Send(ctx, msg)

	receiverUserInfo, exist, err := repo.UserRepo.GetByUserID(ctx, questionUserID)
	if err != nil {
		log.Error(err)
		return
	}
	if !exist {
		log.Warnf("user %s not found", questionUserID)
		return
	}

	externalNotificationMsg := &schema.ExternalNotificationMsg{
		ReceiverUserID: receiverUserInfo.ID,
		ReceiverEmail:  receiverUserInfo.EMail,
		ReceiverLang:   receiverUserInfo.Language,
	}
	rawData := &schema.NewAnswerTemplateRawData{
		QuestionTitle:   questionTitle,
		QuestionID:      questionID,
		AnswerID:        answerID,
		AnswerSummary:   answerSummary,
		UnsubscribeCode: token.GenerateToken(),
	}
	answerUser, _, _ := services.UserCommon.GetUserBasicInfoByID(ctx, answerUserID)
	if answerUser != nil {
		rawData.AnswerUserDisplayName = answerUser.DisplayName
	}
	externalNotificationMsg.NewAnswerTemplateRawData = rawData
	services.ExternalNotificationQueueService.Send(ctx, externalNotificationMsg)
}
