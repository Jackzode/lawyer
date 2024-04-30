package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
)

func RegisterQuestionApi(r *gin.RouterGroup) {

	c := controller.QuestionController{}
	// question
	r.GET("/question/info", c.GetQuestion)
	r.GET("/question/invite", c.GetQuestionInviteUserInfo)
	r.GET("/question/page", c.QuestionPage)
	r.GET("/question/similar/tag", c.SimilarQuestion)
	r.GET("/personal/qa/top", c.UserTop)
	r.GET("/personal/question/page", c.PersonalQuestionPage)
	r.GET("/answer/page", c.AdminAnswerPage)
	// question
	r.GET("/personal/collection/page", c.PersonalCollectionPage)
	r.POST("/question", c.AddQuestion)
	r.POST("/question/answer", c.AddQuestionByAnswer)
	r.PUT("/question", c.UpdateQuestion)
	r.PUT("/question/invite", c.UpdateQuestionInviteUser)
	r.DELETE("/question", c.RemoveQuestion)
	r.PUT("/question/status", c.CloseQuestion)
	r.PUT("/question/operation", c.OperationQuestion)
	r.PUT("/question/reopen", c.ReopenQuestion)
	r.GET("/question/similar", c.GetSimilarQuestions)
	r.POST("/question/recover", c.QuestionRecover)
	r.GET("/question/page", c.AdminQuestionPage)
	r.PUT("/question/status", c.AdminUpdateQuestionStatus)
	r.GET("/personal/answer/page", c.PersonalAnswerPage)

}
