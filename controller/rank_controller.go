package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/middleware"
	"github.com/lawyer/service"
)

// RankController rank controller
type RankController struct {
	rankService *service.RankService
}

// NewRankController new controller
func NewRankController(
	rankService *service.RankService) *RankController {
	return &RankController{rankService: rankService}
}

// GetRankPersonalWithPage user personal rank list
// @Summary user personal rank list
// @Description user personal rank list
// @Tags Rank
// @Produce json
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Param username query string false "username"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetRankPersonalPageResp}}
// @Router /answer/api/v1/personal/rank/page [get]
func (cc *RankController) GetRankPersonalWithPage(ctx *gin.Context) {
	req := &schema.GetRankPersonalWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := cc.rankService.GetRankPersonalPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
