package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/middleware"
	"github.com/lawyer/service"
)

type PermissionController struct {
	rankService *service.RankService
}

// NewPermissionController new language controller.
func NewPermissionController(rankService *service.RankService) *PermissionController {
	return &PermissionController{rankService: rankService}
}

// GetPermission check user permission
// @Summary check user permission
// @Description check user permission
// @Tags Permission
// @Security ApiKeyAuth
// @Param Authorization header string true "access-token"
// @Produce json
// @Param action query string true "permission key" Enums(question.add, question.edit, question.edit_without_review, question.delete, question.close, question.reopen, question.vote_up, question.vote_down, question.pin, question.unpin, question.hide, question.show, answer.add, answer.edit, answer.edit_without_review, answer.delete, answer.accept, answer.vote_up, answer.vote_down, answer.invite_someone_to_answer, comment.add, comment.edit, comment.delete, comment.vote_up, comment.vote_down, report.add, tag.add, tag.edit, tag.edit_slug_name, tag.edit_without_review, tag.delete, tag.synonym, link.url_limit, vote.detail, answer.audit, question.audit, tag.audit, tag.use_reserved_tag)
// @Success 200 {object} handler.RespBody{data=map[string]bool}
// @Router /answer/api/v1/permission [get]
func (u *PermissionController) GetPermission(ctx *gin.Context) {
	req := &schema.GetPermissionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	userID := middleware.GetLoginUserIDFromContext(ctx)
	ops, requireRanks, err := u.rankService.CheckOperationPermissionsForRanks(ctx, userID, req.Actions)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	lang := utils.GetLangByCtx(ctx)
	mapping := make(map[string]*schema.GetPermissionResp, len(ops))
	for i, action := range req.Actions {
		t := &schema.GetPermissionResp{HasPermission: ops[i]}
		t.TrTip(lang, requireRanks[i])
		mapping[action] = t
	}
	handler.HandleResponse(ctx, err, mapping)
}
