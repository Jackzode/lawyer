package templaterender

import (
	"context"
	services "github.com/lawyer/initServer/initServices"

	"github.com/lawyer/commons/schema"
)

func (t *TemplateRenderController) AnswerList(ctx context.Context, req *schema.AnswerListReq) ([]*schema.AnswerInfo, int64, error) {
	return services.AnswerService.SearchList(ctx, req)
}
