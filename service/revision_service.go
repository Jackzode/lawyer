package service

import (
	"context"
	"encoding/json"
	constant2 "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity "github.com/lawyer/commons/entity"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/repo"
	"time"

	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/obj"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// RevisionService user service
type RevisionService struct {
}

func NewRevisionService() *RevisionService {
	return &RevisionService{}
}

func (rs *RevisionService) RevisionAudit(ctx context.Context, req *schema.RevisionAuditReq) (err error) {
	revisioninfo, exist, err := repo.RevisionRepo.GetRevisionByID(ctx, req.ID)
	if err != nil {
		return
	}
	if !exist {
		return
	}
	if revisioninfo.Status != entity.RevisionUnreviewedStatus {
		return
	}
	if req.Operation == schema.RevisionAuditReject {
		err = repo.RevisionRepo.UpdateStatus(ctx, req.ID, entity.RevisionReviewRejectStatus, req.UserID)
		return
	}
	if req.Operation == schema.RevisionAuditApprove {
		objectType, objectTypeerr := obj.GetObjectTypeStrByObjectID(revisioninfo.ObjectID)
		if objectTypeerr != nil {
			return objectTypeerr
		}
		revisionitem := &schema.GetRevisionResp{}
		_ = copier.Copy(revisionitem, revisioninfo)
		rs.parseItem(ctx, revisionitem)
		var saveErr error
		switch objectType {
		case constant2.QuestionObjectType:
			if !req.CanReviewQuestion {
				saveErr = errors.BadRequest(reason.RevisionNoPermission)
			} else {
				saveErr = rs.revisionAuditQuestion(ctx, revisionitem)
			}
		case constant2.AnswerObjectType:
			if !req.CanReviewAnswer {
				saveErr = errors.BadRequest(reason.RevisionNoPermission)
			} else {
				saveErr = rs.revisionAuditAnswer(ctx, revisionitem)
			}
		case constant2.TagObjectType:
			if !req.CanReviewTag {
				saveErr = errors.BadRequest(reason.RevisionNoPermission)
			} else {
				saveErr = rs.revisionAuditTag(ctx, revisionitem)
			}
		}
		if saveErr != nil {
			return saveErr
		}
		err = repo.RevisionRepo.UpdateStatus(ctx, req.ID, entity.RevisionReviewPassStatus, req.UserID)
		return
	}

	return nil
}

func (rs *RevisionService) revisionAuditQuestion(ctx context.Context, revisionitem *schema.GetRevisionResp) (err error) {
	questioninfo, ok := revisionitem.ContentParsed.(*schema.QuestionInfo)
	if ok {
		var PostUpdateTime time.Time
		dbquestion, exist, dberr := repo.QuestionRepo.GetQuestion(ctx, questioninfo.ID)
		if dberr != nil || !exist {
			return
		}

		PostUpdateTime = time.Unix(questioninfo.UpdateTime, 0)
		if dbquestion.PostUpdateTime.Unix() > PostUpdateTime.Unix() {
			PostUpdateTime = dbquestion.PostUpdateTime
		}
		question := &entity.Question{}
		question.ID = questioninfo.ID
		question.Title = questioninfo.Title
		question.OriginalText = questioninfo.Content
		question.ParsedText = questioninfo.HTML
		question.UpdatedAt = time.Unix(questioninfo.UpdateTime, 0)
		question.PostUpdateTime = PostUpdateTime
		question.LastEditUserID = revisionitem.UserID
		saveerr := repo.QuestionRepo.UpdateQuestion(ctx, question, []string{"title", "original_text", "parsed_text", "updated_at", "post_update_time", "last_edit_user_id"})
		if saveerr != nil {
			return saveerr
		}
		objectTagTags := make([]*schema.TagItem, 0)
		for _, tag := range questioninfo.Tags {
			item := &schema.TagItem{}
			item.SlugName = tag.SlugName
			objectTagTags = append(objectTagTags, item)
		}
		objectTagData := schema.TagChange{}
		objectTagData.ObjectID = question.ID
		objectTagData.Tags = objectTagTags
		saveerr = services.TagCommonService.ObjectChangeTag(ctx, &objectTagData)
		if saveerr != nil {
			return saveerr
		}
		services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           revisionitem.UserID,
			ObjectID:         revisionitem.ObjectID,
			ActivityTypeKey:  constant2.ActQuestionEdited,
			RevisionID:       revisionitem.ID,
			OriginalObjectID: revisionitem.ObjectID,
		})
	}
	return nil
}

func (rs *RevisionService) revisionAuditAnswer(ctx context.Context, revisionitem *schema.GetRevisionResp) (err error) {
	answerinfo, ok := revisionitem.ContentParsed.(*schema.AnswerInfo)
	if ok {

		var PostUpdateTime time.Time
		dbquestion, exist, dberr := repo.QuestionRepo.GetQuestion(ctx, answerinfo.QuestionID)
		if dberr != nil || !exist {
			return
		}

		PostUpdateTime = time.Unix(answerinfo.UpdateTime, 0)
		if dbquestion.PostUpdateTime.Unix() > PostUpdateTime.Unix() {
			PostUpdateTime = dbquestion.PostUpdateTime
		}

		insertData := new(entity.Answer)
		insertData.ID = answerinfo.ID
		insertData.OriginalText = answerinfo.Content
		insertData.ParsedText = answerinfo.HTML
		insertData.UpdatedAt = time.Unix(answerinfo.UpdateTime, 0)
		insertData.LastEditUserID = revisionitem.UserID
		saveerr := repo.AnswerRepo.UpdateAnswer(ctx, insertData, []string{"original_text", "parsed_text", "updated_at", "last_edit_user_id"})
		if saveerr != nil {
			return saveerr
		}
		saveerr = services.QuestionCommon.UpdatePostSetTime(ctx, answerinfo.QuestionID, PostUpdateTime)
		if saveerr != nil {
			return saveerr
		}
		questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, answerinfo.QuestionID)
		if err != nil {
			return err
		}
		if !exist {
			return errors.BadRequest(reason.QuestionNotFound)
		}
		msg := &schema.NotificationMsg{
			TriggerUserID:  revisionitem.UserID,
			ReceiverUserID: questionInfo.UserID,
			Type:           schema.NotificationTypeInbox,
			ObjectID:       answerinfo.ID,
		}
		msg.ObjectType = constant2.AnswerObjectType
		msg.NotificationAction = constant2.NotificationUpdateAnswer
		services.NotificationQueueService.Send(ctx, msg)

		services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           revisionitem.UserID,
			ObjectID:         insertData.ID,
			OriginalObjectID: insertData.ID,
			ActivityTypeKey:  constant2.ActAnswerEdited,
			RevisionID:       revisionitem.ID,
		})
	}
	return nil
}

func (rs *RevisionService) revisionAuditTag(ctx context.Context, revisionitem *schema.GetRevisionResp) (err error) {
	taginfo, ok := revisionitem.ContentParsed.(*schema.GetTagResp)
	if ok {
		tag := &entity.Tag{}
		tag.ID = taginfo.TagID
		tag.OriginalText = taginfo.OriginalText
		tag.ParsedText = taginfo.ParsedText
		saveerr := repo.TagRepo.UpdateTag(ctx, tag)
		if saveerr != nil {
			return saveerr
		}

		tagInfo, exist, err := services.TagCommonService.GetTagByID(ctx, taginfo.TagID)
		if err != nil {
			return err
		}
		if !exist {
			return errors.BadRequest(reason.TagNotFound)
		}
		if tagInfo.MainTagID == 0 && len(tagInfo.SlugName) > 0 {
			log.Debugf("tag %s update slug_name", tagInfo.SlugName)
			tagList, err := repo.TagRepo.GetTagList(ctx, &entity.Tag{MainTagID: converter.StringToInt64(tagInfo.ID)})
			if err != nil {
				return err
			}
			updateTagSlugNames := make([]string, 0)
			for _, tag := range tagList {
				updateTagSlugNames = append(updateTagSlugNames, tag.SlugName)
			}
			err = repo.TagRepo.UpdateTagSynonym(ctx, updateTagSlugNames, converter.StringToInt64(tagInfo.ID), tagInfo.MainTagSlugName)
			if err != nil {
				return err
			}
		}

		services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           revisionitem.UserID,
			ObjectID:         taginfo.TagID,
			OriginalObjectID: taginfo.TagID,
			ActivityTypeKey:  constant2.ActTagEdited,
			RevisionID:       revisionitem.ID,
		})
	}
	return nil
}

// GetUnreviewedRevisionPage get unreviewed list
func (rs *RevisionService) GetUnreviewedRevisionPage(ctx context.Context, req *schema.RevisionSearch) (
	resp *pager.PageModel, err error) {
	revisionResp := make([]*schema.GetUnreviewedRevisionResp, 0)
	if len(req.GetCanReviewObjectTypes()) == 0 {
		return pager.NewPageModel(0, revisionResp), nil
	}
	revisionPage, total, err := repo.RevisionRepo.GetUnreviewedRevisionPage(
		ctx, req.Page, 1, req.GetCanReviewObjectTypes())
	if err != nil {
		return nil, err
	}
	for _, rev := range revisionPage {
		item := &schema.GetUnreviewedRevisionResp{}
		_, ok := constant2.ObjectTypeNumberMapping[rev.ObjectType]
		if !ok {
			continue
		}
		item.Type = constant2.ObjectTypeNumberMapping[rev.ObjectType]
		info, err := services.ObjService.GetUnreviewedRevisionInfo(ctx, rev.ObjectID)
		if err != nil {
			return nil, err
		}
		item.Info = info
		revisionitem := &schema.GetRevisionResp{}
		_ = copier.Copy(revisionitem, rev)
		rs.parseItem(ctx, revisionitem)
		item.UnreviewedInfo = revisionitem

		// get user info
		userInfo, exists, e := services.UserCommon.GetUserBasicInfoByID(ctx, revisionitem.UserID)
		if e != nil {
			return nil, e
		}
		if exists {
			var uinfo schema.UserBasicInfo
			_ = copier.Copy(&uinfo, userInfo)
			item.UnreviewedInfo.UserInfo = uinfo
		}
		revisionResp = append(revisionResp, item)
	}
	return pager.NewPageModel(total, revisionResp), nil
}

// GetRevisionList get revision list all
func (rs *RevisionService) GetRevisionList(ctx context.Context, req *schema.GetRevisionListReq) (resp []schema.GetRevisionResp, err error) {
	var (
		rev  entity.Revision
		revs []entity.Revision
	)

	resp = []schema.GetRevisionResp{}
	_ = copier.Copy(&rev, req)

	revs, err = repo.RevisionRepo.GetRevisionList(ctx, &rev)
	if err != nil {
		return
	}

	for _, r := range revs {
		var (
			uinfo schema.UserBasicInfo
			item  schema.GetRevisionResp
		)

		_ = copier.Copy(&item, r)
		rs.parseItem(ctx, &item)

		// get user info
		userInfo, exists, e := services.UserCommon.GetUserBasicInfoByID(ctx, item.UserID)
		if e != nil {
			return nil, e
		}
		if exists {
			err = copier.Copy(&uinfo, userInfo)
			item.UserInfo = uinfo
		}
		resp = append(resp, item)
	}
	return
}

func (rs *RevisionService) parseItem(ctx context.Context, item *schema.GetRevisionResp) {
	var (
		err          error
		question     entity.QuestionWithTagsRevision
		questionInfo *schema.QuestionInfo
		answer       entity.Answer
		answerInfo   *schema.AnswerInfo
		tag          entity.Tag
		tagInfo      *schema.GetTagResp
	)

	switch item.ObjectType {
	case constant2.ObjectTypeStrMapping["question"]:
		err = json.Unmarshal([]byte(item.Content), &question)
		if err != nil {
			break
		}
		questionInfo = services.QuestionCommon.ShowFormatWithTag(ctx, &question)
		item.ContentParsed = questionInfo
	case constant2.ObjectTypeStrMapping["answer"]:
		err = json.Unmarshal([]byte(item.Content), &answer)
		if err != nil {
			break
		}
		answerInfo = services.AnswerService.ShowFormat(ctx, &answer)
		item.ContentParsed = answerInfo
	case constant2.ObjectTypeStrMapping["tag"]:
		err = json.Unmarshal([]byte(item.Content), &tag)
		if err != nil {
			break
		}
		tagInfo = &schema.GetTagResp{
			TagID:         tag.ID,
			CreatedAt:     tag.CreatedAt.Unix(),
			UpdatedAt:     tag.UpdatedAt.Unix(),
			SlugName:      tag.SlugName,
			DisplayName:   tag.DisplayName,
			OriginalText:  tag.OriginalText,
			ParsedText:    tag.ParsedText,
			FollowCount:   tag.FollowCount,
			QuestionCount: tag.QuestionCount,
			Recommend:     tag.Recommend,
			Reserved:      tag.Reserved,
		}
		tagInfo.GetExcerpt()
		item.ContentParsed = tagInfo
	}

	if err != nil {
		item.ContentParsed = item.Content
	}
	item.CreatedAtParsed = item.CreatedAt.Unix()
}

// CheckCanUpdateRevision can check revision
func (rs *RevisionService) CheckCanUpdateRevision(ctx context.Context, req *schema.CheckCanQuestionUpdate) (
	resp *schema.ErrTypeData, err error) {
	_, exist, err := repo.RevisionRepo.ExistUnreviewedByObjectID(ctx, req.ID)
	if err != nil {
		return nil, nil
	}
	if exist {
		return &schema.ErrTypeToast, errors.BadRequest(reason.RevisionReviewUnderway)
	}
	return nil, nil
}
