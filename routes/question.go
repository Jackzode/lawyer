package routes

import (
	"github.com/gin-gonic/gin"
	"lawyer/controller"
)

func InitQuestionApiRoutes(engine *gin.RouterGroup) {
	//need login
	engine.POST("/add", (&controller.QuestionController{}).AddQuestion)

}
