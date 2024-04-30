package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
)

func RegisterCommentApi(r *gin.RouterGroup) {
	c := &controller.CommentController{}
	// comment
	r.POST("/comment", c.AddComment)
	r.DELETE("/comment", c.RemoveComment)
	r.PUT("/comment", c.UpdateComment)
	r.GET("/comment/page", c.GetCommentWithPage)
	r.GET("/personal/comment/page", c.GetCommentPersonalWithPage)
	r.GET("/comment", c.GetComment)
}
