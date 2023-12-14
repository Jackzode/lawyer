package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"lawyer/common"
	"lawyer/dao"
	"lawyer/types"
	"time"
)

type QuestionService struct {
}

func (q *QuestionService) AddQuestion(ctx *gin.Context, req *types.QuestionReq) error {

	//todo 对tag的操作

	question := &types.Question{}
	now := time.Now()
	question.UserID = "0"
	question.Title = req.Title
	question.OriginalText = req.Content
	question.ParsedText = req.HTML
	question.AcceptedAnswerID = "0"
	question.LastAnswerID = "0"
	question.LastEditUserID = "0"
	//question.PostUpdateTime = nil
	question.Status = common.QuestionStatusAvailable
	question.RevisionID = "0"
	question.CreatedAt = now
	question.PostUpdateTime = now
	question.Pin = common.QuestionUnPin
	question.Show = common.QuestionShow
	question.ID = "100000001"
	addQuestion := dao.AddQuestion(ctx, question)
	if addQuestion != 1 {
		fmt.Println("add question failed")
		return errors.New("add question failed")
	}
	return nil
}
