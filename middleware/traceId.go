package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/utils"
)

func TraceId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceId := utils.GetTraceIdFromHeader(ctx)
		if traceId == "" {
			traceId = utils.GenerateTraceId()
		}
		ctx.Set(constant.TraceID, traceId)
		ctx.Next()
	}

}
