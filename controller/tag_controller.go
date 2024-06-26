package controller

import (
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/middleware"
	"github.com/lawyer/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/service/permission"
	"github.com/segmentfault/pacman/errors"
)

// TagController tag controller
type TagController struct{}

// NewTagController new controller
func NewTagController() *TagController {
	return &TagController{}
}

// SearchTagLike get tag list
// @Summary get tag list
// @Description get tag list
// @Tags Tag
// @Produce json
// @Security ApiKeyAuth
// @Param tag query string false "tag"
// @Success 200 {object} handler.RespBody{data=[]schema.GetTagResp}
// @Router /answer/api/v1/question/tags [get]
func (tc *TagController) SearchTagLike(ctx *gin.Context) {
	req := &schema.SearchTagLikeReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	resp, err := service.TagServicer.SearchTagLike(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetTagsBySlugName
// @Summary get tags list
// @Description get tags list
// @Tags Tag
// @Produce json
// @Param tags query []string false "string collection" collectionFormat(csv)
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/tags [get]
func (tc *TagController) GetTagsBySlugName(ctx *gin.Context) {
	req := &schema.SearchTagsBySlugName{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.TagList = strings.Split(req.Tags, ",")
	// req.IsAdmin = middleware.GetIsAdminFromContext(ctx)
	resp, err := service.TagServicer.GetTagsBySlugName(ctx, req.TagList)
	handler.HandleResponse(ctx, err, resp)
}

// RemoveTag delete tag
// @Summary delete tag
// @Description delete tag
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.RemoveTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag [delete]
func (tc *TagController) RemoveTag(ctx *gin.Context) {
	req := &schema.RemoveTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	can, err := service.RankServicer.CheckOperationPermission(ctx, req.UserID, permission.TagDelete, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	err = service.TagServicer.RemoveTag(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// AddTag add tag
// @Summary add tag
// @Description add tag
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.AddTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag [post]
func (tc *TagController) AddTag(ctx *gin.Context) {
	req := &schema.AddTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	//req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.UserID = utils.GetUidFromTokenByCtx(ctx)
	//校验权限
	canList, err := service.RankServicer.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.TagAdd,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !canList[0] {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	//添加逻辑
	resp, err := service.TagServicer.AddTag(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateTag update tag
// @Summary update tag
// @Description update tag
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.UpdateTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag [put]
func (tc *TagController) UpdateTag(ctx *gin.Context) {
	req := &schema.UpdateTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := service.RankServicer.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.TagEdit,
		permission.TagEditWithoutReview,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !canList[0] {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	req.NoNeedReview = canList[1]

	err = service.TagServicer.UpdateTag(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
	} else {
		handler.HandleResponse(ctx, err, &schema.UpdateTagResp{WaitForReview: !req.NoNeedReview})
	}
}

// RecoverTag recover delete tag
// @Summary recover delete tag
// @Description recover delete tag
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.RecoverTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag/recover [post]
func (tc *TagController) RecoverTag(ctx *gin.Context) {
	req := &schema.RecoverTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := service.RankServicer.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.TagUnDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !canList[0] {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = service.TagServicer.RecoverTag(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetTagInfo get tag one
// @Summary get tag one
// @Description get tag one
// @Tags Tag
// @Accept json
// @Produce json
// @Param tag_id query string true "tag id"
// @Param tag_name query string true "tag name"
// @Success 200 {object} handler.RespBody{data=schema.GetTagResp}
// @Router /answer/api/v1/tag [get]
func (tc *TagController) GetTagInfo(ctx *gin.Context) {
	req := &schema.GetTagInfoReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := service.RankServicer.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.TagEdit,
		permission.TagDelete,
		permission.TagUnDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = canList[0]
	req.CanDelete = canList[1]
	req.CanRecover = canList[2]

	resp, err := service.TagServicer.GetTagInfo(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// @Router /answer/api/v1/tags/page [get]
func (tc *TagController) GetTagWithPage(ctx *gin.Context) {
	req := &schema.GetTagWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	//req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.UserID = utils.GetUidFromTokenByCtx(ctx)

	resp, err := service.TagServicer.GetTagWithPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetFollowingTags get following tag list
// @Summary get following tag list
// @Description get following tag list
// @Security ApiKeyAuth
// @Tags Tag
// @Produce json
// @Success 200 {object} handler.RespBody{data=[]schema.GetFollowingTagsResp}
// @Router /answer/api/v1/tags/following [get]
func (tc *TagController) GetFollowingTags(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)
	resp, err := service.TagServicer.GetFollowingTags(ctx, userID)
	handler.HandleResponse(ctx, err, resp)
}

// GetTagSynonyms get tag synonyms
// @Summary get tag synonyms
// @Description get tag synonyms
// @Tags Tag
// @Produce json
// @Param tag_id query int true "tag id"
// @Success 200 {object} handler.RespBody{data=schema.GetTagSynonymsResp}
// @Router /answer/api/v1/tag/synonyms [get]
func (tc *TagController) GetTagSynonyms(ctx *gin.Context) {
	req := &schema.GetTagSynonymsReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	can, err := service.RankServicer.CheckOperationPermission(ctx, req.UserID, permission.TagSynonym, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = can

	resp, err := service.TagServicer.GetTagSynonyms(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateTagSynonym update tag
// @Summary update tag
// @Description update tag
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.UpdateTagSynonymReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag/synonym [put]
func (tc *TagController) UpdateTagSynonym(ctx *gin.Context) {
	req := &schema.UpdateTagSynonymReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	can, err := service.RankServicer.CheckOperationPermission(ctx, req.UserID, permission.TagSynonym, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = service.TagServicer.UpdateTagSynonym(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
