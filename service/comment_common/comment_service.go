package comment_common

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/repo"

	"github.com/lawyer/commons/schema"
	"github.com/segmentfault/pacman/errors"
)

// CommentCommonRepo comment repository
type CommentCommonRepo interface {
	GetComment(ctx context.Context, commentID string) (comment *entity.Comment, exist bool, err error)
	GetCommentCount(ctx context.Context) (count int64, err error)
	RemoveAllUserComment(ctx context.Context, userID string) (err error)
}

// CommentCommonService user service
type CommentCommonService struct {
}

// NewCommentCommonService new comment service
func NewCommentCommonService() *CommentCommonService {
	return &CommentCommonService{}
}

// GetComment get comment one
func (cs *CommentCommonService) GetComment(ctx context.Context, commentID string) (resp *schema.GetCommentResp, err error) {
	comment, exist, err := repo.CommentRepo.GetComment(ctx, commentID)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.BadRequest(reason.CommentNotFound)
	}

	resp = &schema.GetCommentResp{}
	resp.SetFromComment(comment)
	return resp, nil
}
