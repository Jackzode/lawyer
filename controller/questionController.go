package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lawyer/service"
	"lawyer/types"
)

type QuestionController struct {
}

func (q *QuestionController) AddQuestion(ctx *gin.Context) {
	//参数绑定
	req := &types.QuestionReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	//todo 验证是否具有新增问题的权限

	//todo 校验tag

	//新增问题
	questionService := &service.QuestionService{}
	questionService.AddQuestion(ctx, req)

	//response
}
