package service

import (
	"context"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/pkg/obj"
	"github.com/lawyer/pkg/uid"
	"github.com/lawyer/repo"
	"github.com/segmentfault/pacman/errors"
)

// ObjServicer user service
type ObjService struct {
}

// NewObjService new object service
func NewObjService() *ObjService {
	return &ObjService{}
}
func (os *ObjService) GetUnreviewedRevisionInfo(ctx context.Context, objectID string) (objInfo *schema.UnreviewedRevisionInfoInfo, err error) {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return nil, err
	}
	switch objectType {
	case constant.QuestionObjectType:
		questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, objectID)
		if err != nil {
			return nil, err
		}
		if utils.GetEnableShortID(ctx) {
			questionInfo.ID = uid.EnShortID(questionInfo.ID)
		}
		if !exist {
			break
		}
		taglist, err := TagServicer.GetObjectEntityTag(ctx, objectID)
		if err != nil {
			return nil, err
		}
		TagServicer.TagsFormatRecommendAndReserved(ctx, taglist)
		tags, err := schema.TagFormat(taglist)
		if err != nil {
			return nil, err
		}
		objInfo = &schema.UnreviewedRevisionInfoInfo{
			ObjectID: questionInfo.ID,
			Title:    questionInfo.Title,
			Content:  questionInfo.OriginalText,
			Html:     questionInfo.ParsedText,
			Tags:     tags,
		}
	case constant.AnswerObjectType:
		answerInfo, exist, err := repo.AnswerRepo.GetAnswer(ctx, objectID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}

		questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, answerInfo.QuestionID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		if utils.GetEnableShortID(ctx) {
			questionInfo.ID = uid.EnShortID(questionInfo.ID)
		}
		objInfo = &schema.UnreviewedRevisionInfoInfo{
			ObjectID: answerInfo.ID,
			Title:    questionInfo.Title,
			Content:  answerInfo.OriginalText,
			Html:     answerInfo.ParsedText,
		}

	case constant.TagObjectType:
		tagInfo, exist, err := repo.TagRepo.GetTagByID(ctx, objectID, true)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		objInfo = &schema.UnreviewedRevisionInfoInfo{
			ObjectID: tagInfo.ID,
			Title:    tagInfo.SlugName,
			Content:  tagInfo.OriginalText,
			Html:     tagInfo.ParsedText,
		}
	}
	if objInfo == nil {
		err = errors.BadRequest(reason.ObjectNotFound)
	}
	return objInfo, err
}

/*
根据obj id来确定当前请求的是什么，可以是question，answer，comment，tag
如果是question，我们查db获取完成的question信息；
如果是answer，我们需要获取answer信息，question信息
如果是comment，我们要获取comment信息，answer信息，question信息
如果是tag信息，我们获取tag信息即可
*/
func (os *ObjService) GetInfo(ctx context.Context, objectID string) (objInfo *schema.SimpleObjectInfo, err error) {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return nil, err
	}
	switch objectType {
	case constant.QuestionObjectType:
		questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, objectID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		objInfo = &schema.SimpleObjectInfo{
			ObjectID:            questionInfo.ID,
			ObjectCreatorUserID: questionInfo.UserID,
			QuestionID:          questionInfo.ID,
			QuestionStatus:      questionInfo.Status,
			ObjectType:          objectType,
			Title:               questionInfo.Title,
			Content:             questionInfo.ParsedText, // todo trim
		}
	case constant.AnswerObjectType:
		answerInfo, exist, err := repo.AnswerRepo.GetAnswer(ctx, objectID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, answerInfo.QuestionID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		objInfo = &schema.SimpleObjectInfo{
			ObjectID:            answerInfo.ID,
			ObjectCreatorUserID: answerInfo.UserID,
			QuestionID:          answerInfo.QuestionID,
			QuestionStatus:      questionInfo.Status,
			AnswerID:            answerInfo.ID,
			ObjectType:          objectType,
			Title:               questionInfo.Title,    // this should be question title
			Content:             answerInfo.ParsedText, // todo trim
		}
	case constant.CommentObjectType:
		commentInfo, exist, err := repo.CommentCommonRepo.GetComment(ctx, objectID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		objInfo = &schema.SimpleObjectInfo{
			ObjectID:            commentInfo.ID,
			ObjectCreatorUserID: commentInfo.UserID,
			ObjectType:          objectType,
			Content:             commentInfo.ParsedText, // todo trim
			CommentID:           commentInfo.ID,
		}
		if len(commentInfo.QuestionID) > 0 {
			questionInfo, exist, err := repo.QuestionRepo.GetQuestion(ctx, commentInfo.QuestionID)
			if err != nil {
				return nil, err
			}
			if exist {
				objInfo.QuestionID = questionInfo.ID
				objInfo.QuestionStatus = questionInfo.Status
				objInfo.Title = questionInfo.Title
			}
			answerInfo, exist, err := repo.AnswerRepo.GetAnswer(ctx, commentInfo.ObjectID)
			if err != nil {
				return nil, err
			}
			if exist {
				objInfo.AnswerID = answerInfo.ID
			}
		}
	case constant.TagObjectType:
		tagInfo, exist, err := repo.TagRepo.GetTagByID(ctx, objectID, true)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		objInfo = &schema.SimpleObjectInfo{
			ObjectID:   tagInfo.ID,
			TagID:      tagInfo.ID,
			ObjectType: objectType,
			Title:      tagInfo.ParsedText,
			Content:    tagInfo.ParsedText, // todo trim
		}
	}
	if objInfo == nil {
		err = errors.BadRequest(reason.ObjectNotFound)
	}
	return objInfo, err
}
