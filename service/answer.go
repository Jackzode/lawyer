package service

import (
	"context"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/repo"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/htmltext"
	"github.com/lawyer/pkg/uid"
)

type AnswerRepo interface {
	AddAnswer(ctx context.Context, answer *entity.Answer) (err error)
	RemoveAnswer(ctx context.Context, id string) (err error)
	RecoverAnswer(ctx context.Context, answerID string) (err error)
	UpdateAnswer(ctx context.Context, answer *entity.Answer, cols []string) (err error)
	GetAnswer(ctx context.Context, id string) (answer *entity.Answer, exist bool, err error)
	GetAnswerList(ctx context.Context, answer *entity.Answer) (answerList []*entity.Answer, err error)
	GetAnswerPage(ctx context.Context, page, pageSize int, answer *entity.Answer) (answerList []*entity.Answer, total int64, err error)
	UpdateAcceptedStatus(ctx context.Context, acceptedAnswerID string, questionID string) error
	GetByID(ctx context.Context, answerID string) (*entity.Answer, bool, error)
	GetCountByQuestionID(ctx context.Context, questionID string) (int64, error)
	GetCountByUserID(ctx context.Context, userID string) (int64, error)
	GetIDsByUserIDAndQuestionID(ctx context.Context, userID string, questionID string) ([]string, error)
	SearchList(ctx context.Context, search *entity.AnswerSearch) ([]*entity.Answer, int64, error)
	AdminSearchList(ctx context.Context, search *schema.AdminAnswerPageReq) ([]*entity.Answer, int64, error)
	UpdateAnswerStatus(ctx context.Context, answerID string, status int) (err error)
	GetAnswerCount(ctx context.Context) (count int64, err error)
	RemoveAllUserAnswer(ctx context.Context, userID string) (err error)
}

// AnswerCommonServicer user service
type AnswerCommon struct {
}

func NewAnswerCommon() *AnswerCommon {
	return &AnswerCommon{}
}

func (as *AnswerCommon) SearchAnswerIDs(ctx context.Context, userID, questionID string) ([]string, error) {
	ids, err := repo.AnswerRepo.GetIDsByUserIDAndQuestionID(ctx, userID, questionID)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (as *AnswerCommon) AdminSearchList(ctx context.Context, req *schema.AdminAnswerPageReq) (
	resp []*entity.Answer, count int64, err error) {
	resp, count, err = repo.AnswerRepo.AdminSearchList(ctx, req)
	if utils.GetEnableShortID(ctx) {
		for _, item := range resp {
			item.ID = uid.EnShortID(item.ID)
			item.QuestionID = uid.EnShortID(item.QuestionID)
		}
	}
	return resp, count, err
}

func (as *AnswerCommon) Search(ctx context.Context, search *entity.AnswerSearch) ([]*entity.Answer, int64, error) {
	list, count, err := repo.AnswerRepo.SearchList(ctx, search)
	if err != nil {
		return list, count, err
	}
	return list, count, err
}

// ShowFormat 做了一个结构体转换
func (as *AnswerCommon) ShowFormat(ctx context.Context, data *entity.Answer) *schema.AnswerInfo {
	info := schema.AnswerInfo{}
	info.ID = data.ID
	info.QuestionID = data.QuestionID
	info.Content = data.OriginalText
	info.HTML = data.ParsedText
	info.Accepted = data.Accepted
	info.VoteCount = data.VoteCount
	info.CreateTime = data.CreatedAt.Unix()
	info.UpdateTime = data.UpdatedAt.Unix()
	if data.UpdatedAt.Unix() < 1 {
		info.UpdateTime = 0
	}
	info.UserID = data.UserID
	info.UpdateUserID = data.LastEditUserID
	info.Status = data.Status
	info.MemberActions = make([]*schema.PermissionMemberAction, 0)
	return &info
}

func (as *AnswerCommon) AdminShowFormat(ctx context.Context, data *entity.Answer) *schema.AdminAnswerInfo {
	info := schema.AdminAnswerInfo{}
	info.ID = data.ID
	info.QuestionID = data.QuestionID
	info.Accepted = data.Accepted
	info.VoteCount = data.VoteCount
	info.CreateTime = data.CreatedAt.Unix()
	info.UpdateTime = data.UpdatedAt.Unix()
	if data.UpdatedAt.Unix() < 1 {
		info.UpdateTime = 0
	}
	info.UserID = data.UserID
	info.UpdateUserID = data.LastEditUserID
	info.Description = htmltext.FetchExcerpt(data.ParsedText, "...", 240)
	return &info
}
