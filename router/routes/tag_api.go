package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
	"github.com/lawyer/middleware"
)

/*
关于tag，或者说对问题添加话题，我这里打算这样做：
不允许用户新增话题，只能选择目前已经有的话题，只允许管理员添加，
那么暴露给普通用户的接口就只有get tag就行了。
以后可以离线生成tag灌库
*/
func RegisterTagApi(r *gin.RouterGroup) {
	c := controller.NewTagController()
	rg := r.Group("/tag", middleware.AccessToken())
	// tag
	rg.GET("/page", c.GetTagWithPage)
	rg.GET("/following", c.GetFollowingTags)
	rg.GET("/info", c.GetTagInfo)
	rg.GET("/getinfobyslug", c.GetTagsBySlugName)
	rg.GET("/synonyms", c.GetTagSynonyms)
	rg.GET("/question", c.SearchTagLike)
	r.POST("/recover", c.RecoverTag)
	r.PUT("/update/synonym", c.UpdateTagSynonym)

	//admin接口，先不梳理
	rg.POST("/add", c.AddTag)
	rg.PUT("/update", c.UpdateTag)
	rg.DELETE("/del", c.RemoveTag)
}
