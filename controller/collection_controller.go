package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/middleware"
	"github.com/lawyer/pkg/uid"
	"github.com/lawyer/service"
)

// CollectionController collection controller
type CollectionController struct {
	collectionService *service.CollectionService
}

// NewCollectionController new controller
func NewCollectionController(collectionService *service.CollectionService) *CollectionController {
	return &CollectionController{collectionService: collectionService}
}

// CollectionSwitch add collection
// @Summary add collection
// @Description add collection
// @Tags Collection
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.CollectionSwitchReq true "collection"
// @Success 200 {object} handler.RespBody{data=schema.CollectionSwitchResp}
// @Router /answer/api/v1/collection/switch [post]
func (cc *CollectionController) CollectionSwitch(ctx *gin.Context) {
	req := &schema.CollectionSwitchReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.ObjectID = uid.DeShortID(req.ObjectID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := cc.collectionService.CollectionSwitch(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
