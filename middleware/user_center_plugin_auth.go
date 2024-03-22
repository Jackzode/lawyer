package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/plugin"
	"github.com/segmentfault/pacman/errors"
)

// BanAPIForUserCenter ban api for user center
func BanAPIForUserCenter(ctx *gin.Context) {
	uc, ok := plugin.GetUserCenter()
	if !ok {
		return
	}
	if !uc.Description().EnabledOriginalUserSystem {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		ctx.Abort()
		return
	}
	ctx.Next()
}
