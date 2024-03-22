package templaterender

import (
	"github.com/lawyer/commons/schema"
	"golang.org/x/net/context"
)

func (q *TemplateRenderController) UserInfo(ctx context.Context, req *schema.GetOtherUserInfoByUsernameReq) (resp *schema.GetOtherUserInfoByUsernameResp, err error) {
	return q.userService.GetOtherUserInfoByUsername(ctx, req.Username)
}
