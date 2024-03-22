package comment

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"

	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/service/comment"
	"github.com/lawyer/service/comment_common"
	"github.com/segmentfault/pacman/errors"
)

// commentRepo comment repository
type commentRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewCommentRepo new repository
func NewCommentRepo() comment.CommentRepo {
	return &commentRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// NewCommentCommonRepo new repository
func NewCommentCommonRepo() comment_common.CommentCommonRepo {
	return &commentRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// AddComment add comment
func (cr *commentRepo) AddComment(ctx context.Context, comment *entity.Comment) (err error) {
	comment.ID, err = repo.UniqueIDRepo.GenUniqueIDStr(ctx, comment.TableName())
	if err != nil {
		return err
	}
	_, err = cr.DB.Context(ctx).Insert(comment)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveComment delete comment
func (cr *commentRepo) RemoveComment(ctx context.Context, commentID string) (err error) {
	session := cr.DB.Context(ctx).ID(commentID)
	_, err = session.Update(&entity.Comment{Status: entity.CommentStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateCommentContent update comment
func (cr *commentRepo) UpdateCommentContent(
	ctx context.Context, commentID string, originalText string, parsedText string) (err error) {
	_, err = cr.DB.Context(ctx).ID(commentID).Update(&entity.Comment{
		OriginalText: originalText,
		ParsedText:   parsedText,
	})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetComment get comment one
func (cr *commentRepo) GetComment(ctx context.Context, commentID string) (
	comment *entity.Comment, exist bool, err error) {
	comment = &entity.Comment{}
	exist, err = cr.DB.Context(ctx).ID(commentID).Get(comment)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (cr *commentRepo) GetCommentCount(ctx context.Context) (count int64, err error) {
	list := make([]*entity.Comment, 0)
	count, err = cr.DB.Context(ctx).Where("status = ?", entity.CommentStatusAvailable).FindAndCount(&list)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCommentPage get comment page
func (cr *commentRepo) GetCommentPage(ctx context.Context, commentQuery *comment.CommentQuery) (
	commentList []*entity.Comment, total int64, err error,
) {
	commentList = make([]*entity.Comment, 0)

	session := cr.DB.Context(ctx)
	session.OrderBy(commentQuery.GetOrderBy())
	session.Where("status = ?", entity.CommentStatusAvailable)

	cond := &entity.Comment{ObjectID: commentQuery.ObjectID, UserID: commentQuery.UserID}
	total, err = pager.Help(commentQuery.Page, commentQuery.PageSize, &commentList, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveAllUserComment remove all user comment
func (cr *commentRepo) RemoveAllUserComment(ctx context.Context, userID string) (err error) {
	session := cr.DB.Context(ctx).Where("user_id = ?", userID)
	session.Where("status != ?", entity.CommentStatusDeleted)
	affected, err := session.Update(&entity.Comment{Status: entity.CommentStatusDeleted})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	log.Infof("delete user comment, userID: %s, affected: %d", userID, affected)
	return
}
