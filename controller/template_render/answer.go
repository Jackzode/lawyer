package templaterender

import (
	"context"
	services "github.com/lawyer/service"

	"github.com/lawyer/commons/schema"
)

func (t *TemplateRenderController) AnswerList(ctx context.Context, req *schema.AnswerListReq) ([]*schema.AnswerInfo, int64, error) {
	return services.AnswerServicer.SearchList(ctx, req)
}
