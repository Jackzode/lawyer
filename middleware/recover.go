package middleware

import (
	"github.com/gin-gonic/gin"
)

func RecoverPanic() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

			}
		}()
		//next step
		ctx.Next()
	}
}
