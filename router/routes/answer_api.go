package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
)

func RegisterAnswerApi(r *gin.RouterGroup) {

	c := &controller.AnswerController{}
	// answer
	r.GET("/answer/info", c.Get)
	r.GET("/answer/page", c.AnswerList)
	r.POST("/answer", c.Add)
	r.PUT("/answer", c.Update)
	r.POST("/answer/acceptance", c.Accepted)
	r.DELETE("/answer", c.RemoveAnswer)
	r.POST("/answer/recover", c.RecoverAnswer)
	r.PUT("/answer/status", c.AdminUpdateAnswerStatus)

}
