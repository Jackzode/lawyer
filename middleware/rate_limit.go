package middleware

import (
	"github.com/gin-gonic/gin"
)

type RateLimitMiddleware struct {
	//limitRepo *limit.LimitRepo
}

// NewRateLimitMiddleware new rate limit middleware
func NewRateLimitMiddleware(
// limitRepo *limit.LimitRepo
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		//limitRepo: limitRepo,
	}
}

// DuplicateRequestRejection detects and rejects duplicate requests
// 检查一下当前用户，当前内容是否在五分钟内请求过这个接口，也就是避免重复提交
// It only works for the requests that post content. Such as add question, add answer, comment etc.
func (rm *RateLimitMiddleware) DuplicateRequestRejection(ctx *gin.Context, req any) (reject bool, key string) {
	//userID := GetLoginUserIDFromContext(ctx)
	//fullPath := ctx.FullPath()
	//reqJson, _ := json.Marshal(req)
	//key = encryption.MD5(fmt.Sprintf("%s:%s:%s", userID, fullPath, string(reqJson)))
	//var err error
	//
	//reject, err = rm.limitRepo.CheckAndRecord(ctx, key)
	//if err != nil {
	//	log.Errorf("check and record rate limit error: %s", err.Error())
	//	return false, key
	//}
	//if !reject {
	//	return false, key
	//}
	//log.Debugf("duplicate request: [%s] %s", fullPath, string(reqJson))
	//handler.HandleResponse(ctx, errors.BadRequest(reason.DuplicateRequestError), nil)
	//return true, key
	return true, key

}

// DuplicateRequestClear clear duplicate request record
func (rm *RateLimitMiddleware) DuplicateRequestClear(ctx *gin.Context, key string) {
	//err := rm.limitRepo.ClearRecord(ctx, key)
	//if err != nil {
	//	log.Errorf("clear rate limit error: %s", err.Error())
	//}
}
