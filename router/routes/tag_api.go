package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
)

func RegisterTagApi(r *gin.RouterGroup) {
	c := controller.NewTagController()
	// tag
	r.GET("/tags/page", c.GetTagWithPage)
	r.GET("/tags/following", c.GetFollowingTags)
	r.GET("/tag", c.GetTagInfo)
	r.GET("/tags", c.GetTagsBySlugName)
	r.GET("/tag/synonyms", c.GetTagSynonyms)
	// tag
	r.GET("/question/tags", c.SearchTagLike)
	r.POST("/tag", c.AddTag)
	r.PUT("/tag", c.UpdateTag)
	r.POST("/tag/recover", c.RecoverTag)
	r.DELETE("/tag", c.RemoveTag)
	r.PUT("/tag/synonym", c.UpdateTagSynonym)
}
