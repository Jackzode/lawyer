package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
)

// revision
func RegisterRevisionApi(r *gin.RouterGroup) {
	c := controller.NewRevisionController(nil, nil)
	r.GET("/revisions", c.GetRevisionList)
	r.GET("/revisions/unreviewed", c.GetUnreviewedRevisionList)
	r.PUT("/revisions/audit", c.RevisionAudit)
	r.GET("/revisions/edit/check", c.CheckCanUpdateRevision)

}
