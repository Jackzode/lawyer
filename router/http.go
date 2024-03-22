package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lawyer/middleware"
	"github.com/lawyer/plugin"
)

// NewHTTPServer new http server.
func NewHTTPServer(debug bool,
	answerRouter *AnswerAPIRouter,
	authUserMiddleware *middleware.AuthUserMiddleware,
	avatarMiddleware *middleware.AvatarMiddleware,
	shortIDMiddleware *middleware.ShortIDMiddleware,
	templateRouter *TemplateRouter,
	pluginAPIRouter *PluginAPIRouter,
) *gin.Engine {

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.ExtractAndSetAcceptLanguage, shortIDMiddleware.SetShortIDFlag())
	//健康检测
	r.GET("/health", func(ctx *gin.Context) { ctx.String(200, "OK") })

	//html, _ := fs.Sub(ui.Template, "template")
	//htmlTemplate := template.Must(template.New("").Funcs(funcMap).ParseFS(html, "*"))
	//r.SetHTMLTemplate(htmlTemplate)
	//r.Use(middleware.HeadersByRequestURI())
	//viewRouter.Register(r)

	rootGroup := r.Group("")
	//static := r.Group("")
	//static.Use(avatarMiddleware.AvatarThumb(), authUserMiddleware.VisitAuth())
	//staticRouter.RegisterStaticRouter(static)

	// do not need to login
	mustUnAuthV1 := r.Group("/lawyer/api")
	answerRouter.RegisterMustUnAuthAnswerAPIRouter(authUserMiddleware, mustUnAuthV1)

	// register api that no need to login
	unAuthV1 := r.Group("/lawyer/api")
	unAuthV1.Use(authUserMiddleware.Auth(), authUserMiddleware.EjectUserBySiteInfo())
	answerRouter.RegisterUnAuthAnswerAPIRouter(unAuthV1)

	// register api that must be authenticated
	authV1 := r.Group("/lawyer/api")
	authV1.Use(authUserMiddleware.MustAuth())
	answerRouter.RegisterAnswerAPIRouter(authV1)

	adminauthV1 := r.Group("/lawyer/admin")
	adminauthV1.Use(authUserMiddleware.AdminAuth())
	answerRouter.RegisterAnswerAdminAPIRouter(adminauthV1)

	templateRouter.RegisterTemplateRouter(rootGroup)

	// plugin routes
	pluginAPIRouter.RegisterUnAuthConnectorRouter(mustUnAuthV1)
	pluginAPIRouter.RegisterAuthUserConnectorRouter(authV1)
	pluginAPIRouter.RegisterAuthAdminConnectorRouter(adminauthV1)

	_ = plugin.CallAgent(func(agent plugin.Agent) error {
		agent.RegisterUnAuthRouter(mustUnAuthV1)
		agent.RegisterAuthUserRouter(authV1)
		agent.RegisterAuthAdminRouter(adminauthV1)
		return nil
	})
	return r
}
