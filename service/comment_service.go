package service

import (
	"context"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/pkg/token"
	"github.com/lawyer/repo"
	"github.com/lawyer/service/permission"
	"time"

	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/htmltext"
	"github.com/lawyer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// CommentRepo comment repository
type CommentRepo interface {
	AddComment(ctx context.Context, comment *entity.Comment) (err error)
	RemoveComment(ctx context.Context, commentID string) (err error)
	UpdateCommentContent(ctx context.Context, commentID string, original string, parsedText string) (err error)
	GetComment(ctx context.Context, commentID string) (comment *entity.Comment, exist bool, err error)
	GetCommentPage(ctx context.Context, commentQuery *utils.CommentQuery) (
		comments []*entity.Comment, total int64, err error)
}

// CommentServicer user service
type CommentService struct {
}

// NewCommentService new comment service
func NewCommentService() *CommentService {
	return &CommentService{}
}

// AddComment add comment
func (cs *CommentService) AddComment(ctx context.Context, req *schema.AddCommentReq) (
	resp *schema.GetCommentResp, err error) {
	comment := &entity.Comment{}
	_ = copier.Copy(comment, req)
	comment.Status = entity.CommentStatusAvailable

	// add question id
	objInfo, err := ObjServicer.GetInfo(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}
	objInfo.ObjectID = uid.DeShortID(objInfo.ObjectID)
	objInfo.QuestionID = uid.DeShortID(objInfo.QuestionID)
	objInfo.AnswerID = uid.DeShortID(objInfo.AnswerID)
	if objInfo.ObjectType == constant.QuestionObjectType || objInfo.ObjectType == constant.AnswerObjectType {
		comment.QuestionID = objInfo.QuestionID
	}

	if len(req.ReplyCommentID) > 0 {
		replyComment, exist, err := repo.CommentCommonRepo.GetComment(ctx, req.ReplyCommentID)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.BadRequest(reason.CommentNotFound)
		}
		comment.SetReplyUserID(replyComment.UserID)
		comment.SetReplyCommentID(replyComment.ID)
	} else {
		comment.SetReplyUserID("")
		comment.SetReplyCommentID("")
	}

	err = repo.CommentRepo.AddComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	resp = &schema.GetCommentResp{}
	resp.SetFromComment(comment)
	resp.MemberActions = permission.GetCommentPermission(ctx, req.UserID, resp.UserID,
		time.Now(), req.CanEdit, req.CanDelete)

	commentResp, err := cs.addCommentNotification(ctx, req, resp, comment, objInfo)
	if err != nil {
		return commentResp, err
	}

	// get user info
	userInfo, exist, err := UserCommonServicer.GetUserBasicInfoByID(ctx, resp.UserID)
	if err != nil {
		return nil, err
	}
	if exist {
		resp.Username = userInfo.Username
		resp.UserDisplayName = userInfo.DisplayName
		resp.UserAvatar = userInfo.Avatar
		resp.UserStatus = userInfo.Status
	}

	activityMsg := &schema.ActivityMsg{
		UserID:           comment.UserID,
		ObjectID:         comment.ID,
		OriginalObjectID: req.ObjectID,
		ActivityTypeKey:  constant.ActQuestionCommented,
	}
	switch objInfo.ObjectType {
	case constant.QuestionObjectType:
		activityMsg.ActivityTypeKey = constant.ActQuestionCommented
	case constant.AnswerObjectType:
		activityMsg.ActivityTypeKey = constant.ActAnswerCommented
	}
	ActivityQueueServicer.Send(ctx, activityMsg)
	return resp, nil
}

func (cs *CommentService) addCommentNotification(
	ctx context.Context, req *schema.AddCommentReq, resp *schema.GetCommentResp,
	comment *entity.Comment, objInfo *schema.SimpleObjectInfo) (*schema.GetCommentResp, error) {
	// The priority of the notification
	// 1. reply to user
	// 2. comment mention to user
	// 3. answer or question was commented
	alreadyNotifiedUserID := make(map[string]bool)

	// get reply user info
	if len(resp.ReplyUserID) > 0 && resp.ReplyUserID != req.UserID {
		replyUser, exist, err := UserCommonServicer.GetUserBasicInfoByID(ctx, resp.ReplyUserID)
		if err != nil {
			return nil, err
		}
		if exist {
			resp.ReplyUsername = replyUser.Username
			resp.ReplyUserDisplayName = replyUser.DisplayName
			resp.ReplyUserStatus = replyUser.Status
		}
		cs.notificationCommentReply(ctx, replyUser.ID, comment.ID, req.UserID)
		alreadyNotifiedUserID[replyUser.ID] = true
		return nil, nil
	}

	if len(req.MentionUsernameList) > 0 {
		alreadyNotifiedUserIDs := cs.notificationMention(
			ctx, req.MentionUsernameList, comment.ID, req.UserID, alreadyNotifiedUserID)
		for _, userID := range alreadyNotifiedUserIDs {
			alreadyNotifiedUserID[userID] = true
		}
		return nil, nil
	}

	if objInfo.ObjectType == constant.QuestionObjectType && !alreadyNotifiedUserID[objInfo.ObjectCreatorUserID] {
		cs.notificationQuestionComment(ctx, objInfo.ObjectCreatorUserID,
			objInfo.QuestionID, objInfo.Title, comment.ID, req.UserID, comment.OriginalText)
	} else if objInfo.ObjectType == constant.AnswerObjectType && !alreadyNotifiedUserID[objInfo.ObjectCreatorUserID] {
		cs.notificationAnswerComment(ctx, objInfo.QuestionID, objInfo.Title, objInfo.AnswerID,
			objInfo.ObjectCreatorUserID, comment.ID, req.UserID, comment.OriginalText)
	}
	return nil, nil
}

// RemoveComment delete comment
func (cs *CommentService) RemoveComment(ctx context.Context, req *schema.RemoveCommentReq) (err error) {
	return repo.CommentRepo.RemoveComment(ctx, req.CommentID)
}

// UpdateComment update comment
func (cs *CommentService) UpdateComment(ctx context.Context, req *schema.UpdateCommentReq) (
	resp *schema.UpdateCommentResp, err error) {
	old, exist, err := repo.CommentCommonRepo.GetComment(ctx, req.CommentID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.CommentNotFound)
	}
	// user can't edit the comment that was posted by others except admin
	if !req.IsAdmin && req.UserID != old.UserID {
		return nil, errors.BadRequest(reason.CommentNotFound)
	}

	// user can edit the comment that was posted by himself before deadline.
	// admin can edit it at any time
	if !req.IsAdmin && (time.Now().After(old.CreatedAt.Add(constant.CommentEditDeadline))) {
		return nil, errors.BadRequest(reason.CommentCannotEditAfterDeadline)
	}

	if err = repo.CommentRepo.UpdateCommentContent(ctx, old.ID, req.OriginalText, req.ParsedText); err != nil {
		return nil, err
	}
	resp = &schema.UpdateCommentResp{
		CommentID:    old.ID,
		OriginalText: req.OriginalText,
		ParsedText:   req.ParsedText,
	}
	return resp, nil
}

// GetComment get comment one
func (cs *CommentService) GetComment(ctx context.Context, req *schema.GetCommentReq) (resp *schema.GetCommentResp, err error) {
	comment, exist, err := repo.CommentCommonRepo.GetComment(ctx, req.ID)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.BadRequest(reason.CommentNotFound)
	}

	resp = &schema.GetCommentResp{
		CommentID:      comment.ID,
		CreatedAt:      comment.CreatedAt.Unix(),
		UserID:         comment.UserID,
		ReplyUserID:    comment.GetReplyUserID(),
		ReplyCommentID: comment.GetReplyCommentID(),
		ObjectID:       comment.ObjectID,
		VoteCount:      comment.VoteCount,
		OriginalText:   comment.OriginalText,
		ParsedText:     comment.ParsedText,
	}

	// get comment user info
	if len(resp.UserID) > 0 {
		commentUser, exist, err := UserCommonServicer.GetUserBasicInfoByID(ctx, resp.UserID)
		if err != nil {
			return nil, err
		}
		if exist {
			resp.Username = commentUser.Username
			resp.UserDisplayName = commentUser.DisplayName
			resp.UserAvatar = commentUser.Avatar
			resp.UserStatus = commentUser.Status
		}
	}

	// get reply user info
	if len(resp.ReplyUserID) > 0 {
		replyUser, exist, err := UserCommonServicer.GetUserBasicInfoByID(ctx, resp.ReplyUserID)
		if err != nil {
			return nil, err
		}
		if exist {
			resp.ReplyUsername = replyUser.Username
			resp.ReplyUserDisplayName = replyUser.DisplayName
			resp.ReplyUserStatus = replyUser.Status
		}
	}

	// check if current user vote this comment
	resp.IsVote = cs.checkIsVote(ctx, req.UserID, resp.CommentID)

	resp.MemberActions = permission.GetCommentPermission(ctx, req.UserID, resp.UserID,
		comment.CreatedAt, req.CanEdit, req.CanDelete)
	return resp, nil
}

// GetCommentWithPage get comment list page
func (cs *CommentService) GetCommentWithPage(ctx context.Context, req *schema.GetCommentWithPageReq) (
	pageModel *pager.PageModel, err error) {
	dto := &utils.CommentQuery{
		PageCond:  pager.PageCond{Page: req.Page, PageSize: req.PageSize},
		ObjectID:  req.ObjectID,
		QueryCond: req.QueryCond,
	}
	commentList, total, err := repo.CommentRepo.GetCommentPage(ctx, dto)
	if err != nil {
		return nil, err
	}
	resp := make([]*schema.GetCommentResp, 0)
	for _, comment := range commentList {
		commentResp, err := cs.convertCommentEntity2Resp(ctx, req, comment)
		if err != nil {
			return nil, err
		}
		resp = append(resp, commentResp)
	}

	// if user request the specific comment, add it if not exist.
	if len(req.CommentID) > 0 {
		commentExist := false
		for _, t := range resp {
			if t.CommentID == req.CommentID {
				commentExist = true
				break
			}
		}
		if !commentExist {
			comment, exist, err := repo.CommentCommonRepo.GetComment(ctx, req.CommentID)
			if err != nil {
				return nil, err
			}
			if exist && comment.ObjectID == req.ObjectID {
				commentResp, err := cs.convertCommentEntity2Resp(ctx, req, comment)
				if err != nil {
					return nil, err
				}
				resp = append(resp, commentResp)
			}
		}
	}
	return pager.NewPageModel(total, resp), nil
}

func (cs *CommentService) convertCommentEntity2Resp(ctx context.Context, req *schema.GetCommentWithPageReq,
	comment *entity.Comment) (commentResp *schema.GetCommentResp, err error) {
	commentResp = &schema.GetCommentResp{
		CommentID:      comment.ID,
		CreatedAt:      comment.CreatedAt.Unix(),
		UserID:         comment.UserID,
		ReplyUserID:    comment.GetReplyUserID(),
		ReplyCommentID: comment.GetReplyCommentID(),
		ObjectID:       comment.ObjectID,
		VoteCount:      comment.VoteCount,
		OriginalText:   comment.OriginalText,
		ParsedText:     comment.ParsedText,
	}

	// get comment user info
	if len(commentResp.UserID) > 0 {
		commentUser, exist, err := UserCommonServicer.GetUserBasicInfoByID(ctx, commentResp.UserID)
		if err != nil {
			return nil, err
		}
		if exist {
			commentResp.Username = commentUser.Username
			commentResp.UserDisplayName = commentUser.DisplayName
			commentResp.UserAvatar = commentUser.Avatar
			commentResp.UserStatus = commentUser.Status
		}
	}

	// get reply user info
	if len(commentResp.ReplyUserID) > 0 {
		replyUser, exist, err := UserCommonServicer.GetUserBasicInfoByID(ctx, commentResp.ReplyUserID)
		if err != nil {
			return nil, err
		}
		if exist {
			commentResp.ReplyUsername = replyUser.Username
			commentResp.ReplyUserDisplayName = replyUser.DisplayName
			commentResp.ReplyUserStatus = replyUser.Status
		}
	}

	// check if current user vote this comment
	commentResp.IsVote = cs.checkIsVote(ctx, req.UserID, commentResp.CommentID)

	commentResp.MemberActions = permission.GetCommentPermission(ctx,
		req.UserID, commentResp.UserID, comment.CreatedAt, req.CanEdit, req.CanDelete)
	return commentResp, nil
}

func (cs *CommentService) checkIsVote(ctx context.Context, userID, commentID string) (isVote bool) {
	status := repo.VoteRepo.GetVoteStatus(ctx, commentID, userID)
	return len(status) > 0
}

// GetCommentPersonalWithPage get personal comment list page
func (cs *CommentService) GetCommentPersonalWithPage(ctx context.Context, req *schema.GetCommentPersonalWithPageReq) (
	pageModel *pager.PageModel, err error) {
	if len(req.Username) > 0 {
		userInfo, exist, err := UserCommonServicer.GetUserBasicInfoByUserName(ctx, req.Username)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.BadRequest(reason.UserNotFound)
		}
		req.UserID = userInfo.ID
	}
	if len(req.UserID) == 0 {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	dto := &utils.CommentQuery{
		PageCond:  pager.PageCond{Page: req.Page, PageSize: req.PageSize},
		UserID:    req.UserID,
		QueryCond: "created_at",
	}
	commentList, total, err := repo.CommentRepo.GetCommentPage(ctx, dto)
	if err != nil {
		return nil, err
	}
	resp := make([]*schema.GetCommentPersonalWithPageResp, 0)
	for _, comment := range commentList {
		commentResp := &schema.GetCommentPersonalWithPageResp{
			CommentID: comment.ID,
			CreatedAt: comment.CreatedAt.Unix(),
			ObjectID:  comment.ObjectID,
			Content:   comment.ParsedText, // todo trim
		}
		if len(comment.ObjectID) > 0 {
			objInfo, err := ObjServicer.GetInfo(ctx, comment.ObjectID)
			if err != nil {
				log.Error(err)
			} else {
				commentResp.ObjectType = objInfo.ObjectType
				commentResp.Title = objInfo.Title
				commentResp.UrlTitle = htmltext.UrlTitle(objInfo.Title)
				commentResp.QuestionID = objInfo.QuestionID
				commentResp.AnswerID = objInfo.AnswerID
				if objInfo.QuestionStatus == entity.QuestionStatusDeleted {
					commentResp.Title = "Deleted question"
				}
			}
		}
		resp = append(resp, commentResp)
	}
	return pager.NewPageModel(total, resp), nil
}

func (cs *CommentService) notificationQuestionComment(ctx context.Context, questionUserID,
	questionID, questionTitle, commentID, commentUserID, commentSummary string) {
	if questionUserID == commentUserID {
		return
	}
	// send internal notification
	msg := &schema.NotificationMsg{
		ReceiverUserID: questionUserID,
		TriggerUserID:  commentUserID,
		Type:           schema.NotificationTypeInbox,
		ObjectID:       commentID,
	}
	msg.ObjectType = constant.CommentObjectType
	msg.NotificationAction = constant.NotificationCommentQuestion
	NotificationQueueService.Send(ctx, msg)

	// send external notification
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
	rawData := &schema.NewCommentTemplateRawData{
		QuestionTitle:   questionTitle,
		QuestionID:      questionID,
		CommentID:       commentID,
		CommentSummary:  commentSummary,
		UnsubscribeCode: token.GenerateToken(),
	}
	commentUser, _, _ := UserCommonServicer.GetUserBasicInfoByID(ctx, commentUserID)
	if commentUser != nil {
		rawData.CommentUserDisplayName = commentUser.DisplayName
	}
	externalNotificationMsg.NewCommentTemplateRawData = rawData
	ExternalNotificationQueueService.Send(ctx, externalNotificationMsg)
}

func (cs *CommentService) notificationAnswerComment(ctx context.Context,
	questionID, questionTitle, answerID, answerUserID, commentID, commentUserID, commentSummary string) {
	if answerUserID == commentUserID {
		return
	}

	// Send internal notification.
	msg := &schema.NotificationMsg{
		ReceiverUserID: answerUserID,
		TriggerUserID:  commentUserID,
		Type:           schema.NotificationTypeInbox,
		ObjectID:       commentID,
	}
	msg.ObjectType = constant.CommentObjectType
	msg.NotificationAction = constant.NotificationCommentAnswer
	NotificationQueueService.Send(ctx, msg)

	// Send external notification.
	receiverUserInfo, exist, err := repo.UserRepo.GetByUserID(ctx, answerUserID)
	if err != nil {
		log.Error(err)
		return
	}
	if !exist {
		log.Warnf("user %s not found", answerUserID)
		return
	}
	externalNotificationMsg := &schema.ExternalNotificationMsg{
		ReceiverUserID: receiverUserInfo.ID,
		ReceiverEmail:  receiverUserInfo.EMail,
		ReceiverLang:   receiverUserInfo.Language,
	}
	rawData := &schema.NewCommentTemplateRawData{
		QuestionTitle:   questionTitle,
		QuestionID:      questionID,
		AnswerID:        answerID,
		CommentID:       commentID,
		CommentSummary:  commentSummary,
		UnsubscribeCode: token.GenerateToken(),
	}
	commentUser, _, _ := UserCommonServicer.GetUserBasicInfoByID(ctx, commentUserID)
	if commentUser != nil {
		rawData.CommentUserDisplayName = commentUser.DisplayName
	}
	externalNotificationMsg.NewCommentTemplateRawData = rawData
	ExternalNotificationQueueService.Send(ctx, externalNotificationMsg)
}

func (cs *CommentService) notificationCommentReply(ctx context.Context, replyUserID, commentID, commentUserID string) {
	msg := &schema.NotificationMsg{
		ReceiverUserID: replyUserID,
		TriggerUserID:  commentUserID,
		Type:           schema.NotificationTypeInbox,
		ObjectID:       commentID,
	}
	msg.ObjectType = constant.CommentObjectType
	msg.NotificationAction = constant.NotificationReplyToYou
	NotificationQueueService.Send(ctx, msg)
}

func (cs *CommentService) notificationMention(
	ctx context.Context, mentionUsernameList []string, commentID, commentUserID string,
	alreadyNotifiedUserID map[string]bool) (alreadyNotifiedUserIDs []string) {
	for _, username := range mentionUsernameList {
		userInfo, exist, err := UserCommonServicer.GetUserBasicInfoByUserName(ctx, username)
		if err != nil {
			log.Error(err)
			continue
		}
		if exist && !alreadyNotifiedUserID[userInfo.ID] {
			msg := &schema.NotificationMsg{
				ReceiverUserID: userInfo.ID,
				TriggerUserID:  commentUserID,
				Type:           schema.NotificationTypeInbox,
				ObjectID:       commentID,
			}
			msg.ObjectType = constant.CommentObjectType
			msg.NotificationAction = constant.NotificationMentionYou
			NotificationQueueService.Send(ctx, msg)
			alreadyNotifiedUserIDs = append(alreadyNotifiedUserIDs, userInfo.ID)
		}
	}
	return alreadyNotifiedUserIDs
}
