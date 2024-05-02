package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/repo"
	"strings"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/obj"
	"github.com/lawyer/pkg/uid"
)

// ActivityComRepo activity repository
type ActivityRepo interface {
	GetObjectAllActivity(ctx context.Context, objectID string, showVote bool) (activityList []*entity.Activity, err error)
}

// ActivityServicer activity service
type ActivityService struct {
}

// NewActivityService new activity service
func NewActivityService() *ActivityService {
	return &ActivityService{}
}

// GetObjectTimeline get object timeline
func (as *ActivityService) GetObjectTimeline(ctx context.Context, req *schema.GetObjectTimelineReq) (
	resp *schema.GetObjectTimelineResp, err error) {
	resp = &schema.GetObjectTimelineResp{
		ObjectInfo: &schema.ActObjectInfo{},
		Timeline:   make([]*schema.ActObjectTimeline, 0),
	}

	resp.ObjectInfo, err = as.getTimelineMainObjInfo(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}

	activityList, err := repo.ActivityActivityRepo.GetObjectAllActivity(ctx, req.ObjectID, req.ShowVote)
	if err != nil {
		return nil, err
	}
	for _, act := range activityList {
		item := &schema.ActObjectTimeline{
			ActivityID: act.ID,
			RevisionID: converter.IntToString(act.RevisionID),
			CreatedAt:  act.CreatedAt.Unix(),
			Cancelled:  act.Cancelled == entity.ActivityCancelled,
			ObjectID:   act.ObjectID,
			UserInfo:   &schema.UserBasicInfo{},
		}
		item.ObjectType, _ = obj.GetObjectTypeStrByObjectID(act.ObjectID)
		if item.Cancelled {
			item.CancelledAt = act.CancelledAt.Unix()
		}

		if item.ObjectType == constant.QuestionObjectType || item.ObjectType == constant.AnswerObjectType {
			if utils.GetEnableShortID(ctx) {
				item.ObjectID = uid.EnShortID(act.ObjectID)
			}
		}

		cfg, err := utils.GetConfigByID(ctx, act.ActivityType)
		if err != nil {
			glog.Slog.Errorf("fail to get config by id: %d, err: %v, act id is: %s", act.ActivityType, err, act.ID)
		} else {
			// database save activity type is number, change to activity type string is like "question.asked".
			// so we need to cut the front part of '.', only need string like 'asked'
			_, item.ActivityType, _ = strings.Cut(cfg.Key, ".")
			// format activity type string to show
			if isHidden, formattedActivityType := formatActivity(item.ActivityType); isHidden {
				continue
			} else {
				item.ActivityType = formattedActivityType
			}
		}

		// if activity is down vote, only admin can see who does it.
		if item.ActivityType == constant.ActDownVote && !req.IsAdmin {
			item.UserInfo.Username = "N/A"
			item.UserInfo.DisplayName = "N/A"
		} else {
			if act.TriggerUserID > 0 {
				item.UserInfo.ID = fmt.Sprintf("%d", act.TriggerUserID)
			} else {
				item.UserInfo.ID = act.UserID
			}
		}

		item.Comment = as.getTimelineActivityComment(ctx, item.ObjectID, item.ObjectType, item.ActivityType, item.RevisionID)
		resp.Timeline = append(resp.Timeline, item)
	}
	as.formatTimelineUserInfo(ctx, resp.Timeline)
	return
}

func (as *ActivityService) getTimelineMainObjInfo(ctx context.Context, objectID string) (
	resp *schema.ActObjectInfo, err error) {
	resp = &schema.ActObjectInfo{}
	objInfo, err := ObjServicer.GetInfo(ctx, objectID)
	if err != nil {
		return nil, err
	}
	resp.Title = objInfo.Title
	if objInfo.ObjectType == constant.TagObjectType {
		tag, exist, _ := TagCommonServicer.GetTagByID(ctx, objInfo.TagID)
		if exist {
			resp.Title = tag.SlugName
			resp.MainTagSlugName = tag.MainTagSlugName
		}
	}
	resp.ObjectType = objInfo.ObjectType
	resp.QuestionID = objInfo.QuestionID
	resp.AnswerID = objInfo.AnswerID
	if len(objInfo.ObjectCreatorUserID) > 0 {
		// get object creator user info
		userBasicInfo, exist, err := UserCommonServicer.GetUserBasicInfoByID(ctx, objInfo.ObjectCreatorUserID)
		if err != nil {
			return nil, err
		}
		if exist {
			resp.Username = userBasicInfo.Username
			resp.DisplayName = userBasicInfo.DisplayName
		}
	}
	return resp, nil
}

func (as *ActivityService) getTimelineActivityComment(ctx context.Context, objectID, objectType,
	activityType, revisionID string) (comment string) {
	if objectType == constant.CommentObjectType {
		commentInfo, err := CommentCommonService.GetComment(ctx, objectID)
		if err != nil {
			glog.Slog.Error(err)
		} else {
			return commentInfo.ParsedText
		}
		return
	}

	if activityType == constant.ActEdited {
		revision, err := RevisionComServicer.GetRevision(ctx, revisionID)
		if err != nil {
			glog.Slog.Error(err)
		} else {
			return revision.Log
		}
		return
	}
	if activityType == constant.ActClosed {
		// only question can be closed
		metaInfo, err := MetaService.GetMetaByObjectIdAndKey(ctx, objectID, entity.QuestionCloseReasonKey)
		if err != nil {
			glog.Slog.Error(err)
		} else {
			closeMsg := &schema.CloseQuestionMeta{}
			if err := json.Unmarshal([]byte(metaInfo.Value), closeMsg); err == nil {
				return closeMsg.CloseMsg
			}
		}
	}
	return ""
}

func (as *ActivityService) formatTimelineUserInfo(ctx context.Context, timeline []*schema.ActObjectTimeline) {
	userExist := make(map[string]bool)
	userIDs := make([]string, 0)
	for _, info := range timeline {
		if len(info.UserInfo.ID) == 0 || userExist[info.UserInfo.ID] {
			continue
		}
		userIDs = append(userIDs, info.UserInfo.ID)
	}
	if len(userIDs) == 0 {
		return
	}
	userInfoMapping, err := UserCommonServicer.BatchUserBasicInfoByID(ctx, userIDs)
	if err != nil {
		glog.Slog.Error(err)
		return
	}
	for _, info := range timeline {
		if len(info.UserInfo.ID) == 0 {
			continue
		}
		info.UserInfo = userInfoMapping[info.UserInfo.ID]
	}
}

// GetObjectTimelineDetail get object timeline
func (as *ActivityService) GetObjectTimelineDetail(ctx context.Context, req *schema.GetObjectTimelineDetailReq) (
	resp *schema.GetObjectTimelineDetailResp, err error) {
	resp = &schema.GetObjectTimelineDetailResp{}
	resp.OldRevision, _ = as.getOneObjectDetail(ctx, req.OldRevisionID)
	resp.NewRevision, _ = as.getOneObjectDetail(ctx, req.NewRevisionID)
	return resp, nil
}

// GetObjectTimelineDetail get object detail
func (as *ActivityService) getOneObjectDetail(ctx context.Context, revisionID string) (
	resp *schema.ObjectTimelineDetail, err error) {
	resp = &schema.ObjectTimelineDetail{Tags: make([]*schema.ObjectTimelineTag, 0)}

	// if request revision is 0, return null object detail.
	if revisionID == "0" {
		return nil, nil
	}

	revision, err := RevisionComServicer.GetRevision(ctx, revisionID)
	if err != nil {
		glog.Slog.Warn(err)
		return nil, nil
	}
	objInfo, err := ObjServicer.GetInfo(ctx, revision.ObjectID)
	if err != nil {
		return nil, err
	}

	switch objInfo.ObjectType {
	case constant.QuestionObjectType:
		data := &entity.QuestionWithTagsRevision{}
		if err = json.Unmarshal([]byte(revision.Content), data); err != nil {
			glog.Slog.Errorf("revision parsing error %s", err)
			return resp, nil
		}
		for _, tag := range data.Tags {
			resp.Tags = append(resp.Tags, &schema.ObjectTimelineTag{
				SlugName:        tag.SlugName,
				DisplayName:     tag.DisplayName,
				MainTagSlugName: tag.MainTagSlugName,
				Recommend:       tag.Recommend,
				Reserved:        tag.Reserved,
			})
		}
		resp.Title = data.Title
		resp.OriginalText = data.OriginalText
	case constant.AnswerObjectType:
		data := &entity.Answer{}
		if err = json.Unmarshal([]byte(revision.Content), data); err != nil {
			glog.Slog.Errorf("revision parsing error %s", err)
			return resp, nil
		}
		resp.Title = objInfo.Title // answer show question title
		resp.OriginalText = data.OriginalText
	case constant.TagObjectType:
		data := &entity.Tag{}
		if err = json.Unmarshal([]byte(revision.Content), data); err != nil {
			glog.Slog.Errorf("revision parsing error %s", err)
			return resp, nil
		}
		resp.Title = data.DisplayName
		resp.OriginalText = data.OriginalText
		resp.SlugName = data.SlugName
		resp.MainTagSlugName = data.MainTagSlugName
	default:
		glog.Slog.Errorf("unknown object type %s", objInfo.ObjectType)
	}
	return resp, nil
}

func formatActivity(activityType string) (isHidden bool, formattedActivityType string) {
	if activityType == constant.ActVotedUp ||
		activityType == constant.ActVotedDown ||
		activityType == constant.ActFollow {
		return true, ""
	}
	if activityType == constant.ActVoteUp {
		return false, constant.ActUpVote
	}
	if activityType == constant.ActVoteDown {
		return false, constant.ActDownVote
	}
	if activityType == constant.ActAccepted {
		return false, constant.ActAccept
	}
	return false, activityType
}
