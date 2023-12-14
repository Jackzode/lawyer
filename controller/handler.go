package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ResponseHandler(ctx *gin.Context, code int32, errMsg string, data interface{}) {
	response := new(Response)
	response.Code = code
	response.ErrMsg = errMsg
	response.Data = data
	//respStr, err := jsoniter.Marshal(response)
	//if err != nil {
	//	respStr = []byte(common.ResponseErr)
	//}
	ctx.JSON(http.StatusOK, response)
}

type Response struct {
	Code   int32       `json:"code"`
	ErrMsg string      `json:"errMsg"`
	Data   interface{} `json:"data"`
}
