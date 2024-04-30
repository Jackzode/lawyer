package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
)

func RegisterNotificationApi(r *gin.RouterGroup) {
	c := controller.NewNotificationController(nil, nil)
	// notification
	r.GET("/notification/status", c.GetRedDot)
	r.PUT("/notification/status", c.ClearRedDot)
	r.GET("/notification/page", c.GetList)
	r.PUT("/notification/read/state/all", c.ClearUnRead)
	r.PUT("/notification/read/state", c.ClearIDUnRead)
}
