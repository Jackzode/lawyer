package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func RecoverPanic() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		//next step
		ctx.Next()
	}
}
