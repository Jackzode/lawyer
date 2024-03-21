package router

import (
	sc "github.com/apache/incubator-answer/internal/service/service_config"
	"github.com/gin-gonic/gin"
)

// StaticRouter static api router
type StaticRouter struct {
	serviceConfig *sc.ServiceConfig
}

// NewStaticRouter new static api router
func NewStaticRouter(serviceConfig *sc.ServiceConfig) *StaticRouter {
	return &StaticRouter{
		serviceConfig: serviceConfig,
	}
}

// RegisterStaticRouter register static api router
func (a *StaticRouter) RegisterStaticRouter(r *gin.RouterGroup) {
	r.Static("/uploads", a.serviceConfig.UploadPath)
}
