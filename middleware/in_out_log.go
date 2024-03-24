package middleware

import (
	"github.com/gin-gonic/gin"
)

func InOutLog() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//in log

		//next step
		ctx.Next()

		//out log
	}
}
