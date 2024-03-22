package templaterender

import (
	"context"

	"github.com/lawyer/commons/schema"
)

func (t *TemplateRenderController) AnswerList(ctx context.Context, req *schema.AnswerListReq) ([]*schema.AnswerInfo, int64, error) {
	return t.answerService.SearchList(ctx, req)
}
