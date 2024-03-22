package repo_test

import (
	"context"
	"github.com/lawyer/commons/entity"
	"testing"

	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/repo/comment"
	"github.com/lawyer/repo/unique"
	commentService "github.com/lawyer/service/comment"
	"github.com/stretchr/testify/assert"
)

func buildCommentEntity() *entity.Comment {
	return &entity.Comment{
		UserID:       "1",
		ObjectID:     "1",
		QuestionID:   "1",
		VoteCount:    1,
		Status:       entity.CommentStatusAvailable,
		OriginalText: "# title",
		ParsedText:   "<h1>Title</h1>",
	}
}

func Test_commentRepo_AddComment(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()
	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	assert.NoError(t, err)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
}

func Test_commentRepo_GetCommentPage(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()
	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	assert.NoError(t, err)

	resp, total, err := commentRepo.GetCommentPage(context.TODO(), &commentService.CommentQuery{
		PageCond: pager.PageCond{
			Page:     1,
			PageSize: 10,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, total, int64(1))
	assert.Equal(t, resp[0].ID, testCommentEntity.ID)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
}

func Test_commentRepo_UpdateComment(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	commonCommentRepo := comment.NewCommentCommonRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()
	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	assert.NoError(t, err)

	testCommentEntity.ParsedText = "test"
	err = commentRepo.UpdateCommentContent(context.TODO(), testCommentEntity.ID, "test", "test")
	assert.NoError(t, err)

	newComment, exist, err := commonCommentRepo.GetComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testCommentEntity.ParsedText, newComment.ParsedText)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
}
