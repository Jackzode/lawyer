/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package object_info

import (
	"context"
	"github.com/apache/incubator-answer/commons/constant"
	"github.com/apache/incubator-answer/commons/constant/reason"

	"github.com/apache/incubator-answer/commons/schema"
	"github.com/apache/incubator-answer/internal/base/handler"
	answercommon "github.com/apache/incubator-answer/internal/service/answer_common"
	"github.com/apache/incubator-answer/internal/service/comment_common"
	questioncommon "github.com/apache/incubator-answer/internal/service/question_common"
	tagcommon "github.com/apache/incubator-answer/internal/service/tag_common"
	"github.com/apache/incubator-answer/pkg/obj"
	"github.com/apache/incubator-answer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
)

// ObjService user service
type ObjService struct {
	answerRepo   answercommon.AnswerRepo
	questionRepo questioncommon.QuestionRepo
	commentRepo  comment_common.CommentCommonRepo
	tagRepo      tagcommon.TagCommonRepo
	tagCommon    *tagcommon.TagCommonService
}

// NewObjService new object service
func NewObjService(
	answerRepo answercommon.AnswerRepo,
	questionRepo questioncommon.QuestionRepo,
	commentRepo comment_common.CommentCommonRepo,
	tagRepo tagcommon.TagCommonRepo,
	tagCommon *tagcommon.TagCommonService,
) *ObjService {
	return &ObjService{
		answerRepo:   answerRepo,
		questionRepo: questionRepo,
		commentRepo:  commentRepo,
		tagRepo:      tagRepo,
		tagCommon:    tagCommon,
	}
}
func (os *ObjService) GetUnreviewedRevisionInfo(ctx context.Context, objectID string) (objInfo *schema.UnreviewedRevisionInfoInfo, err error) {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return nil, err
	}
	switch objectType {
	case constant.QuestionObjectType:
		questionInfo, exist, err := os.questionRepo.GetQuestion(ctx, objectID)
		if err != nil {
			return nil, err
		}
		if handler.GetEnableShortID(ctx) {
			questionInfo.ID = uid.EnShortID(questionInfo.ID)
		}
		if !exist {
			break
		}
		taglist, err := os.tagCommon.GetObjectEntityTag(ctx, objectID)
		if err != nil {
			return nil, err
		}
		os.tagCommon.TagsFormatRecommendAndReserved(ctx, taglist)
		tags, err := os.tagCommon.TagFormat(ctx, taglist)
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
		answerInfo, exist, err := os.answerRepo.GetAnswer(ctx, objectID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}

		questionInfo, exist, err := os.questionRepo.GetQuestion(ctx, answerInfo.QuestionID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		if handler.GetEnableShortID(ctx) {
			questionInfo.ID = uid.EnShortID(questionInfo.ID)
		}
		objInfo = &schema.UnreviewedRevisionInfoInfo{
			ObjectID: answerInfo.ID,
			Title:    questionInfo.Title,
			Content:  answerInfo.OriginalText,
			Html:     answerInfo.ParsedText,
		}

	case constant.TagObjectType:
		tagInfo, exist, err := os.tagRepo.GetTagByID(ctx, objectID, true)
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

// GetInfo get object simple information
func (os *ObjService) GetInfo(ctx context.Context, objectID string) (objInfo *schema.SimpleObjectInfo, err error) {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return nil, err
	}
	switch objectType {
	case constant.QuestionObjectType:
		questionInfo, exist, err := os.questionRepo.GetQuestion(ctx, objectID)
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
		answerInfo, exist, err := os.answerRepo.GetAnswer(ctx, objectID)
		if err != nil {
			return nil, err
		}
		if !exist {
			break
		}
		questionInfo, exist, err := os.questionRepo.GetQuestion(ctx, answerInfo.QuestionID)
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
		commentInfo, exist, err := os.commentRepo.GetComment(ctx, objectID)
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
			questionInfo, exist, err := os.questionRepo.GetQuestion(ctx, commentInfo.QuestionID)
			if err != nil {
				return nil, err
			}
			if exist {
				objInfo.QuestionID = questionInfo.ID
				objInfo.QuestionStatus = questionInfo.Status
				objInfo.Title = questionInfo.Title
			}
			answerInfo, exist, err := os.answerRepo.GetAnswer(ctx, commentInfo.ObjectID)
			if err != nil {
				return nil, err
			}
			if exist {
				objInfo.AnswerID = answerInfo.ID
			}
		}
	case constant.TagObjectType:
		tagInfo, exist, err := os.tagRepo.GetTagByID(ctx, objectID, true)
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
