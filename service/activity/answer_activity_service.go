package activity

import (
	"context"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/lawyer/service/activity_type"
	"github.com/segmentfault/pacman/log"
)

// AnswerActivityRepo answer activity
type AnswerActivityRepo interface {
	SaveAcceptAnswerActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) (err error)
	SaveCancelAcceptAnswerActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) (err error)
}

// AnswerActivityService answer activity service
type AnswerActivityService struct {
}

// NewAnswerActivityService new comment service
func NewAnswerActivityService() *AnswerActivityService {
	return &AnswerActivityService{}
}

// AcceptAnswer accept answer change activity
func (as *AnswerActivityService) AcceptAnswer(ctx context.Context,
	loginUserID, answerObjID, questionObjID, questionUserID, answerUserID string, isSelf bool) (err error) {
	log.Debugf("user %s want to accept answer %s[%s] for question %s[%s]", loginUserID,
		answerObjID, answerUserID,
		questionObjID, questionUserID)
	operationInfo := as.createAcceptAnswerOperationInfo(ctx, loginUserID,
		answerObjID, questionObjID, questionUserID, answerUserID, isSelf)
	return repo.AnswerActivityRepo.SaveAcceptAnswerActivity(ctx, operationInfo)
}

// CancelAcceptAnswer cancel accept answer change activity
func (as *AnswerActivityService) CancelAcceptAnswer(ctx context.Context,
	loginUserID, answerObjID, questionObjID, questionUserID, answerUserID string) (err error) {
	operationInfo := as.createAcceptAnswerOperationInfo(ctx, loginUserID,
		answerObjID, questionObjID, questionUserID, answerUserID, false)
	return repo.AnswerActivityRepo.SaveCancelAcceptAnswerActivity(ctx, operationInfo)
}

func (as *AnswerActivityService) createAcceptAnswerOperationInfo(ctx context.Context, loginUserID,
	answerObjID, questionObjID, questionUserID, answerUserID string, isSelf bool) *schema.AcceptAnswerOperationInfo {
	operationInfo := &schema.AcceptAnswerOperationInfo{
		TriggerUserID:    loginUserID,
		QuestionObjectID: questionObjID,
		QuestionUserID:   questionUserID,
		AnswerObjectID:   answerObjID,
		AnswerUserID:     answerUserID,
	}
	operationInfo.Activities = as.getActivities(ctx, operationInfo)
	if isSelf {
		for _, activity := range operationInfo.Activities {
			activity.Rank = 0
		}
	}
	return operationInfo
}

func (as *AnswerActivityService) getActivities(ctx context.Context, op *schema.AcceptAnswerOperationInfo) (
	activities []*schema.AcceptAnswerActivity) {
	activities = make([]*schema.AcceptAnswerActivity, 0)

	for _, action := range []string{activity_type.AnswerAccept, activity_type.AnswerAccepted} {
		t := &schema.AcceptAnswerActivity{}
		cfg, err := utils.GetConfigByKey(ctx, action)
		if err != nil {
			log.Warnf("get config by key error: %v", err)
			continue
		}
		t.ActivityType, t.Rank = cfg.ID, cfg.GetIntValue()

		if action == activity_type.AnswerAccept {
			t.ActivityUserID = op.QuestionUserID
			t.TriggerUserID = op.TriggerUserID
			t.OriginalObjectID = op.QuestionObjectID // if activity is 'accept' means this question is accept the answer.
		} else {
			t.ActivityUserID = op.AnswerUserID
			t.TriggerUserID = op.TriggerUserID
			t.OriginalObjectID = op.AnswerObjectID // if activity is 'accepted' means this answer was accepted.
		}
		activities = append(activities, t)
	}
	return activities
}
