package templaterender

import (
	constant "github.com/lawyer/commons/constant"
	services "github.com/lawyer/initServer/initServices"
	"github.com/lawyer/repo"
	"html/template"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/schema"
	"github.com/segmentfault/pacman/log"
)

func (t *TemplateRenderController) Index(ctx *gin.Context, req *schema.QuestionPageReq) ([]*schema.QuestionPageResp, int64, error) {
	return services.QuestionService.GetQuestionPage(ctx, req)
}

func (t *TemplateRenderController) QuestionDetail(ctx *gin.Context, id string) (resp *schema.QuestionInfo, err error) {
	return services.QuestionService.GetQuestion(ctx, id, "", schema.QuestionPermission{})
}

func (t *TemplateRenderController) Sitemap(ctx *gin.Context) {
	general, err := services.SiteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error("get site general failed:", err)
		return
	}
	siteInfo, err := services.SiteInfoCommonService.GetSiteSeo(ctx)
	if err != nil {
		log.Error("get site GetSiteSeo failed:", err)
		return
	}

	questions, err := repo.QuestionRepo.SitemapQuestions(ctx, 1, constant.SitemapMaxSize)
	if err != nil {
		log.Errorf("get sitemap questions failed: %s", err)
		return
	}

	ctx.Header("Content-Type", "application/xml")
	if len(questions) < constant.SitemapMaxSize {
		ctx.HTML(
			http.StatusOK, "sitemap.xml", gin.H{
				"xmlHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
				"list":      questions,
				"general":   general,
				"hastitle": siteInfo.Permalink == constant.PermalinkQuestionIDAndTitle ||
					siteInfo.Permalink == constant.PermalinkQuestionIDAndTitleByShortID,
			},
		)
		return
	}

	questionNum, err := repo.QuestionRepo.GetQuestionCount(ctx)
	if err != nil {
		log.Error("GetQuestionCount error", err)
		return
	}
	var pageList []int
	totalPages := int(math.Ceil(float64(questionNum) / float64(constant.SitemapMaxSize)))
	for i := 1; i <= totalPages; i++ {
		pageList = append(pageList, i)
	}
	ctx.HTML(
		http.StatusOK, "sitemap-list.xml", gin.H{
			"xmlHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"page":      pageList,
			"general":   general,
		},
	)
}

func (t *TemplateRenderController) SitemapPage(ctx *gin.Context, page int) error {
	general, err := services.SiteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error("get site general failed:", err)
		return err
	}
	siteInfo, err := services.SiteInfoCommonService.GetSiteSeo(ctx)
	if err != nil {
		log.Error("get site GetSiteSeo failed:", err)
		return err
	}

	questions, err := repo.QuestionRepo.SitemapQuestions(ctx, page, constant.SitemapMaxSize)
	if err != nil {
		log.Errorf("get sitemap questions failed: %s", err)
		return err
	}
	ctx.Header("Content-Type", "application/xml")
	ctx.HTML(
		http.StatusOK, "sitemap.xml", gin.H{
			"xmlHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"list":      questions,
			"general":   general,
			"hastitle": siteInfo.Permalink == constant.PermalinkQuestionIDAndTitle ||
				siteInfo.Permalink == constant.PermalinkQuestionIDAndTitleByShortID,
		},
	)
	return nil
}
