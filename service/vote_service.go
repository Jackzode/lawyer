package service

import (
	"context"
	"github.com/lawyer/commons/base/translator"
	constant2 "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity2 "github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/utils"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/service/activity_type"
	"github.com/lawyer/service/config"
	"strings"

	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/htmltext"
	"github.com/segmentfault/pacman/log"

	"github.com/lawyer/commons/schema"
	"github.com/segmentfault/pacman/errors"
)

// VoteRepo activity repository
type VoteRepo interface {
	Vote(ctx context.Context, op *schema.VoteOperationInfo) (err error)
	CancelVote(ctx context.Context, op *schema.VoteOperationInfo) (err error)
	GetAndSaveVoteResult(ctx context.Context, objectID, objectType string) (up, down int64, err error)
	ListUserVotes(ctx context.Context, userID string, page int, pageSize int, activityTypes []int) (
		voteList []*entity2.Activity, total int64, err error)
}

// VoteService user service
type VoteService struct {
}

func NewVoteService() *VoteService {
	return &VoteService{}
}

// VoteUp vote up
func (vs *VoteService) VoteUp(ctx context.Context, req *schema.VoteReq) (resp *schema.VoteResp, err error) {
	objectInfo, err := services.ObjService.GetInfo(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}
	// make object id must be decoded
	objectInfo.ObjectID = req.ObjectID

	// check user is voting self or not
	if objectInfo.ObjectCreatorUserID == req.UserID {
		return nil, errors.BadRequest(reason.DisallowVoteYourSelf)
	}

	voteUpOperationInfo := vs.createVoteOperationInfo(ctx, req.UserID, true, objectInfo)

	// vote operation
	if req.IsCancel {
		err = repo.ServiceVoteRepo.CancelVote(ctx, voteUpOperationInfo)
	} else {
		// cancel vote down if exist
		voteOperationInfo := vs.createVoteOperationInfo(ctx, req.UserID, false, objectInfo)
		err = repo.ServiceVoteRepo.CancelVote(ctx, voteOperationInfo)
		if err != nil {
			return nil, err
		}
		err = repo.ServiceVoteRepo.Vote(ctx, voteUpOperationInfo)
	}
	if err != nil {
		return nil, err
	}

	resp = &schema.VoteResp{}
	resp.UpVotes, resp.DownVotes, err = repo.ServiceVoteRepo.GetAndSaveVoteResult(ctx, req.ObjectID, objectInfo.ObjectType)
	if err != nil {
		log.Error(err)
	}
	resp.Votes = resp.UpVotes - resp.DownVotes
	if !req.IsCancel {
		resp.VoteStatus = constant2.ActVoteUp
	}
	return resp, nil
}

// VoteDown vote down
func (vs *VoteService) VoteDown(ctx context.Context, req *schema.VoteReq) (resp *schema.VoteResp, err error) {
	objectInfo, err := services.ObjService.GetInfo(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}
	// make object id must be decoded
	objectInfo.ObjectID = req.ObjectID

	// check user is voting self or not
	if objectInfo.ObjectCreatorUserID == req.UserID {
		return nil, errors.BadRequest(reason.DisallowVoteYourSelf)
	}

	// vote operation
	voteDownOperationInfo := vs.createVoteOperationInfo(ctx, req.UserID, false, objectInfo)
	if req.IsCancel {
		err = repo.ServiceVoteRepo.CancelVote(ctx, voteDownOperationInfo)
		if err != nil {
			return nil, err
		}
	} else {
		// cancel vote up if exist
		err = repo.ServiceVoteRepo.CancelVote(ctx, vs.createVoteOperationInfo(ctx, req.UserID, true, objectInfo))
		if err != nil {
			return nil, err
		}
		err = repo.ServiceVoteRepo.Vote(ctx, voteDownOperationInfo)
		if err != nil {
			return nil, err
		}
	}

	resp = &schema.VoteResp{}
	resp.UpVotes, resp.DownVotes, err = repo.ServiceVoteRepo.GetAndSaveVoteResult(ctx, req.ObjectID, objectInfo.ObjectType)
	if err != nil {
		log.Error(err)
	}
	resp.Votes = resp.UpVotes - resp.DownVotes
	if !req.IsCancel {
		resp.VoteStatus = constant2.ActVoteDown
	}
	return resp, nil
}

// ListUserVotes list user's votes
func (vs *VoteService) ListUserVotes(ctx context.Context, req schema.GetVoteWithPageReq) (resp *pager.PageModel, err error) {
	typeKeys := []string{
		activity_type.QuestionVoteUp,
		activity_type.QuestionVoteDown,
		activity_type.AnswerVoteUp,
		activity_type.AnswerVoteDown,
	}
	activityTypes := make([]int, 0)
	activityTypeMapping := make(map[int]string, 0)

	for _, typeKey := range typeKeys {
		cfg, err := (&config.ConfigService{}).GetConfigByKey(ctx, typeKey)
		if err != nil {
			continue
		}
		activityTypes = append(activityTypes, cfg.ID)
		activityTypeMapping[cfg.ID] = typeKey
	}

	voteList, total, err := repo.ServiceVoteRepo.ListUserVotes(ctx, req.UserID, req.Page, req.PageSize, activityTypes)
	if err != nil {
		return nil, err
	}

	lang := utils.GetLangByCtx(ctx)

	votes := make([]*schema.GetVoteWithPageResp, 0)
	for _, voteInfo := range voteList {
		objInfo, err := services.ObjService.GetInfo(ctx, voteInfo.ObjectID)
		if err != nil {
			log.Error(err)
			continue
		}

		item := &schema.GetVoteWithPageResp{
			CreatedAt:  voteInfo.CreatedAt.Unix(),
			ObjectID:   objInfo.ObjectID,
			QuestionID: objInfo.QuestionID,
			AnswerID:   objInfo.AnswerID,
			ObjectType: objInfo.ObjectType,
			Title:      objInfo.Title,
			UrlTitle:   htmltext.UrlTitle(objInfo.Title),
			Content:    objInfo.Content,
		}
		item.VoteType = translator.Tr(lang,
			activity_type.ActivityTypeFlagMapping[activityTypeMapping[voteInfo.ActivityType]])
		if objInfo.QuestionStatus == entity2.QuestionStatusDeleted {
			item.Title = translator.Tr(lang, constant2.DeletedQuestionTitleTrKey)
		}
		votes = append(votes, item)
	}
	return pager.NewPageModel(total, votes), err
}

func (vs *VoteService) createVoteOperationInfo(ctx context.Context,
	userID string, voteUp bool, objectInfo *schema.SimpleObjectInfo) *schema.VoteOperationInfo {
	// warp vote operation
	voteOperationInfo := &schema.VoteOperationInfo{
		ObjectID:            objectInfo.ObjectID,
		ObjectType:          objectInfo.ObjectType,
		ObjectCreatorUserID: objectInfo.ObjectCreatorUserID,
		OperatingUserID:     userID,
		VoteUp:              voteUp,
		VoteDown:            !voteUp,
	}
	voteOperationInfo.Activities = vs.getActivities(ctx, voteOperationInfo)
	return voteOperationInfo
}

func (vs *VoteService) getActivities(ctx context.Context, op *schema.VoteOperationInfo) (
	activities []*schema.VoteActivity) {
	activities = make([]*schema.VoteActivity, 0)

	var actions []string
	switch op.ObjectType {
	case constant2.QuestionObjectType:
		if op.VoteUp {
			actions = []string{activity_type.QuestionVoteUp, activity_type.QuestionVotedUp}
		} else {
			actions = []string{activity_type.QuestionVoteDown, activity_type.QuestionVotedDown}
		}
	case constant2.AnswerObjectType:
		if op.VoteUp {
			actions = []string{activity_type.AnswerVoteUp, activity_type.AnswerVotedUp}
		} else {
			actions = []string{activity_type.AnswerVoteDown, activity_type.AnswerVotedDown}
		}
	case constant2.CommentObjectType:
		actions = []string{activity_type.CommentVoteUp}
	}

	for _, action := range actions {
		t := &schema.VoteActivity{}
		cfg, err := (&config.ConfigService{}).GetConfigByKey(ctx, action)
		if err != nil {
			log.Warnf("get config by key error: %v", err)
			continue
		}
		t.ActivityType, t.Rank = cfg.ID, cfg.GetIntValue()

		if strings.Contains(action, "voted") {
			t.ActivityUserID = op.ObjectCreatorUserID
			t.TriggerUserID = op.OperatingUserID
		} else {
			t.ActivityUserID = op.OperatingUserID
			t.TriggerUserID = "0"
		}
		activities = append(activities, t)
	}
	return activities
}