package routes

import (
	"github.com/gin-gonic/gin"
	"lawyer/config"
)

func InitRouter() *gin.Engine {
	engine := gin.New()
	gin.SetMode(config.Mode)
	userGroup := engine.Group("/user")
	InitUserApiRouter(userGroup)
	questionGroup := engine.Group("/question")
	InitQuestionApiRoutes(questionGroup)
	return engine
}
