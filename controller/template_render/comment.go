package templaterender

import (
	"context"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
)

func (t *TemplateRenderController) CommentList(
	ctx context.Context,
	objectIDs []string,
) (
	comments map[string][]*schema.GetCommentResp,
	err error,
) {

	comments = make(map[string][]*schema.GetCommentResp, len(objectIDs))

	for _, objectID := range objectIDs {
		var (
			req = &schema.GetCommentWithPageReq{
				Page:      1,
				PageSize:  3,
				ObjectID:  objectID,
				QueryCond: "vote",
				UserID:    "",
			}
			pageModel *pager.PageModel
		)
		pageModel, err = t.commentService.GetCommentWithPage(ctx, req)
		if err != nil {
			return
		}
		li := pageModel.List
		comments[objectID] = li.([]*schema.GetCommentResp)
	}
	return
}
