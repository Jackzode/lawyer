package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
	"github.com/lawyer/middleware"
)

func RegisterQuestionApi(rg *gin.RouterGroup) {

	c := controller.NewQuestionController()
	r := rg.Group("/question", middleware.AccessToken())
	// question
	r.GET("/info", c.GetQuestion)
	r.GET("/invite", c.GetQuestionInviteUserInfo)
	r.GET("/page", c.QuestionPage)
	r.POST("/add", c.AddQuestion)
	r.PUT("/update", c.UpdateQuestion)
	r.PUT("/invite", c.UpdateQuestionInviteUser)
	r.DELETE("/delete", c.RemoveQuestion)
	r.PUT("/status", c.CloseQuestion)
	r.PUT("/operation", c.OperationQuestion)
	r.PUT("/reopen", c.ReopenQuestion)
	r.GET("/similar", c.GetSimilarQuestions)
	r.POST("/recover", c.QuestionRecover)
	r.GET("/similar/tag", c.SimilarQuestion)

	r.GET("/personal/collection/page", c.PersonalCollectionPage)
	r.GET("/personal/question/page", c.PersonalQuestionPage)
	r.GET("/personal/qa/top", c.UserTop)
	r.GET("/personal/answer/page", c.PersonalAnswerPage)
	r.POST("/answer", c.AddQuestionByAnswer)

	//admin
	r.GET("/answer/page", c.AdminAnswerPage)
	r.PUT("/question/status", c.AdminUpdateQuestionStatus)
	//r.GET("/page", c.AdminQuestionPage)
}
