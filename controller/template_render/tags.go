package templaterender

import (
	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"golang.org/x/net/context"
)

func (q *TemplateRenderController) TagList(ctx context.Context, req *schema.GetTagWithPageReq) (resp *pager.PageModel, err error) {
	resp, err = q.tagService.GetTagWithPage(ctx, req)
	if err != nil {
		return
	}
	return
}

func (q *TemplateRenderController) TagInfo(ctx context.Context, req *schema.GetTamplateTagInfoReq) (resp *schema.GetTagResp, questionList []*schema.QuestionPageResp, questionCount int64, err error) {
	dto := &schema.GetTagInfoReq{}
	_ = copier.Copy(dto, req)
	resp, err = q.tagService.GetTagInfo(ctx, dto)
	if err != nil {
		return
	}
	searchQuestion := &schema.QuestionPageReq{}
	searchQuestion.Page = req.Page
	searchQuestion.PageSize = req.PageSize
	searchQuestion.OrderCond = "newest"
	searchQuestion.Tag = req.Name
	searchQuestion.LoginUserID = req.UserID
	questionList, questionCount, err = q.questionService.GetQuestionPage(ctx, searchQuestion)
	if err != nil {
		return
	}
	return resp, questionList, questionCount, err
}
