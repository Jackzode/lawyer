package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/controller"
	templaterender "github.com/lawyer/controller/template_render"
	"github.com/lawyer/controller_admin"
	"github.com/lawyer/middleware"
)

type TemplateRouter struct {
	templateController       *controller.TemplateController
	templateRenderController *templaterender.TemplateRenderController
	siteInfoController       *controller_admin.SiteInfoController
	authUserMiddleware       *middleware.AuthUserMiddleware
}

func NewTemplateRouter(
	templateController *controller.TemplateController,
	templateRenderController *templaterender.TemplateRenderController,
	siteInfoController *controller_admin.SiteInfoController,
	authUserMiddleware *middleware.AuthUserMiddleware,

) *TemplateRouter {
	return &TemplateRouter{
		templateController:       templateController,
		templateRenderController: templateRenderController,
		siteInfoController:       siteInfoController,
		authUserMiddleware:       authUserMiddleware,
	}
}

// RegisterTemplateRouter template router
func (a *TemplateRouter) RegisterTemplateRouter(r *gin.RouterGroup) {
	r.GET("/sitemap.xml", a.templateController.Sitemap)
	r.GET("/sitemap/:page", a.templateController.SitemapPage)

	r.GET("/robots.txt", a.siteInfoController.GetRobots)
	r.GET("/custom.css", a.siteInfoController.GetCss)

	r.GET("/404", a.templateController.Page404)

	//todo add middleware
	seo := r.Group("")
	seo.Use(a.authUserMiddleware.CheckPrivateMode())
	seo.GET("/", a.templateController.Index)
	seo.GET("/questions", a.templateController.QuestionList)
	seo.GET("/questions/:id", a.templateController.QuestionInfo)
	seo.GET("/questions/:id/:title", a.templateController.QuestionInfo)
	seo.GET("/questions/:id/:title/:answerid", a.templateController.QuestionInfo)
	seo.GET("/tags", a.templateController.TagList)
	seo.GET("/tags/:tag", a.templateController.TagInfo)
	seo.GET("/users/:username", a.templateController.UserInfo)
}
