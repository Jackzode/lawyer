package templaterender

import (
	"github.com/lawyer/commons/schema"
	services "github.com/lawyer/service"
	"golang.org/x/net/context"
)

func (q *TemplateRenderController) UserInfo(ctx context.Context, req *schema.GetOtherUserInfoByUsernameReq) (resp *schema.GetOtherUserInfoByUsernameResp, err error) {
	return services.UserServicer.GetOtherUserInfoByUsername(ctx, req.Username)
}
