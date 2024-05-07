package controller

/*

import (
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/base/handler"
	constant "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/utils/checker"
	services "github.com/lawyer/initServer/initServices"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/schema"
	templaterender "github.com/lawyer/controller/template_render"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/htmltext"
	"github.com/lawyer/pkg/obj"
	"github.com/lawyer/pkg/uid"
	"github.com/segmentfault/pacman/log"
)

var SiteUrl = ""

type TemplateController struct {
	scriptPath string
	cssPath    string
	//templateRenderController *templaterender.TemplateRenderController
	//siteInfoService          siteinfo_common.SiteInfoCommonService
}

// NewTemplateController new controller
func NewTemplateController() *TemplateController {
	script, css := GetStyle()
	return &TemplateController{
		scriptPath: script,
		cssPath:    css,
	}
}
func GetStyle() (script, css string) {
	//file, err := ui.Build.ReadFile("build/index.html")
	//if err != nil {
	//	return
	//}
	//scriptRegexp := regexp.MustCompile(`<script defer="defer" src="(.*)"></script>`)
	//scriptData := scriptRegexp.FindStringSubmatch(string(file))
	//cssRegexp := regexp.MustCompile(`<link href="(.*)" rel="stylesheet">`)
	//cssListData := cssRegexp.FindStringSubmatch(string(file))
	//if len(scriptData) == 2 {
	//	script = scriptData[1]
	//}
	//if len(cssListData) == 2 {
	//	css = cssListData[1]
	//}
	return
}
func (tc *TemplateController) SiteInfo(ctx *gin.Context) *schema.TemplateSiteInfoResp {
	var err error
	resp := &schema.TemplateSiteInfoResp{}
	resp.General, err = services.SiteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error(err)
	}
	SiteUrl = resp.General.SiteUrl
	resp.Interface, err = services.SiteInfoService.GetSiteInterface(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Branding, err = services.SiteInfoService.GetSiteBranding(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.SiteSeo, err = services.SiteInfoCommonService.GetSiteSeo(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.CustomCssHtml, err = services.SiteInfoService.GetSiteCustomCssHTML(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.Year = fmt.Sprintf("%d", time.Now().Year())
	return resp
}

// Index question list
func (tc *TemplateController) Index(ctx *gin.Context) {
	req := &schema.QuestionPageReq{
		OrderCond: "newest",
	}
	if handler.BindAndCheck(ctx, req) {
		tc.Page404(ctx)
		return
	}

	var page = req.Page

	data, count, err := tc.templateRenderController.Index(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}

	site := tc.SiteInfo(ctx)
	site.Canonical = site.General.SiteUrl

	UrlUseTitle := false
	if site.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitle {
		UrlUseTitle = true
	}
	site.Title = ""
	tc.html(ctx, http.StatusOK, "question.html", site, gin.H{
		"data":     data,
		"useTitle": UrlUseTitle,
		"page":     templaterender.Paginator(page, req.PageSize, count),
		"path":     "questions",
	})
}

func (tc *TemplateController) QuestionList(ctx *gin.Context) {
	req := &schema.QuestionPageReq{
		OrderCond: "newest",
	}
	if handler.BindAndCheck(ctx, req) {
		tc.Page404(ctx)
		return
	}
	var page = req.Page
	data, count, err := tc.templateRenderController.Index(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	site := tc.SiteInfo(ctx)
	site.Canonical = fmt.Sprintf("%s/questions", site.General.SiteUrl)
	if page > 1 {
		site.Canonical = fmt.Sprintf("%s/questions?page=%d", site.General.SiteUrl, page)
	}

	UrlUseTitle := false
	if site.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitle {
		UrlUseTitle = true
	}
	site.Title = fmt.Sprintf("Questions - %s", site.General.Name)
	tc.html(ctx, http.StatusOK, "question.html", site, gin.H{
		"data":     data,
		"useTitle": UrlUseTitle,
		"page":     templaterender.Paginator(page, req.PageSize, count),
	})
}

func (tc *TemplateController) QuestionInfoeRdirect(ctx *gin.Context, site *schema.TemplateSiteInfoResp, correctTitle bool) (jump bool, url string) {
	questionID := ctx.Param("id")
	title := ctx.Param("title")
	answerID := uid.DeShortID(title)
	titleIsAnswerID := false
	needChangeShortID := false

	objectType, err := obj.GetObjectTypeStrByObjectID(answerID)
	if err == nil && objectType == constant.AnswerObjectType {
		titleIsAnswerID = true
	}

	siteSeo, err := services.SiteInfoService.GetSiteSeo(ctx)
	if err != nil {
		return false, ""
	}
	isShortID := uid.IsShortID(questionID)
	if siteSeo.IsShortLink() {
		if !isShortID {
			questionID = uid.EnShortID(questionID)
			needChangeShortID = true
		}
		if titleIsAnswerID {
			answerID = uid.EnShortID(answerID)
		}
	} else {
		if isShortID {
			needChangeShortID = true
			questionID = uid.DeShortID(questionID)
		}
		if titleIsAnswerID {
			answerID = uid.DeShortID(answerID)
		}
	}

	url = fmt.Sprintf("%s/questions/%s", site.General.SiteUrl, questionID)
	if site.SiteSeo.Permalink == constant.PermalinkQuestionID || site.SiteSeo.Permalink == constant.PermalinkQuestionIDByShortID {
		if len(ctx.Request.URL.Query()) > 0 {
			url = fmt.Sprintf("%s?%s", url, ctx.Request.URL.RawQuery)
		}
		if needChangeShortID {
			return true, url
		}
		//not have title
		if titleIsAnswerID || len(title) == 0 {
			return false, ""
		}

		return true, url
	} else {

		detail, err := tc.templateRenderController.QuestionDetail(ctx, questionID)
		if err != nil {
			tc.Page404(ctx)
			return
		}
		url = fmt.Sprintf("%s/%s", url, htmltext.UrlTitle(detail.Title))
		if titleIsAnswerID {
			url = fmt.Sprintf("%s/%s", url, answerID)
		}

		if len(ctx.Request.URL.Query()) > 0 {
			url = fmt.Sprintf("%s?%s", url, ctx.Request.URL.RawQuery)
		}
		//have title
		if len(title) > 0 && !titleIsAnswerID && correctTitle {
			if needChangeShortID {
				return true, url
			}
			return false, ""
		}
		return true, url
	}
}

// QuestionInfo question and answers info
func (tc *TemplateController) QuestionInfo(ctx *gin.Context) {
	id := ctx.Param("id")
	title := ctx.Param("title")
	answerid := ctx.Param("answerid")
	if checker.IsQuestionsIgnorePath(id) {
		// if id == "ask" {
		//file, err := ui.Build.ReadFile("build/index.html")
		//if err != nil {
		//	log.Error(err)
		//	tc.Page404(ctx)
		//	return
		//}
		ctx.Header("content-type", "text/html;charset=utf-8")
		//ctx.String(http.StatusOK, string(file))
		return
	}

	correctTitle := false

	detail, err := tc.templateRenderController.QuestionDetail(ctx, id)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	encodeTitle := htmltext.UrlTitle(detail.Title)
	if encodeTitle == title {
		correctTitle = true
	}

	site := tc.SiteInfo(ctx)
	jump, jumpurl := tc.QuestionInfoeRdirect(ctx, site, correctTitle)
	if jump {
		ctx.Redirect(http.StatusFound, jumpurl)
		return
	}

	// answers
	answerReq := &schema.AnswerListReq{
		QuestionID: id,
		Order:      "",
		Page:       1,
		PageSize:   999,
		UserID:     "",
	}
	answers, answerCount, err := tc.templateRenderController.AnswerList(ctx, answerReq)
	if err != nil {
		tc.Page404(ctx)
		return
	}

	// comments

	objectIDs := []string{uid.DeShortID(id)}
	for _, answer := range answers {
		answerID := uid.DeShortID(answer.ID)
		objectIDs = append(objectIDs, answerID)
	}
	comments, err := tc.templateRenderController.CommentList(ctx, objectIDs)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	site.Canonical = fmt.Sprintf("%s/questions/%s/%s", site.General.SiteUrl, id, encodeTitle)
	if site.SiteSeo.Permalink == constant.PermalinkQuestionID || site.SiteSeo.Permalink == constant.PermalinkQuestionIDByShortID {
		site.Canonical = fmt.Sprintf("%s/questions/%s", site.General.SiteUrl, id)
	}
	jsonLD := &schema.QAPageJsonLD{}
	jsonLD.Context = "https://schema.org"
	jsonLD.Type = "QAPage"
	jsonLD.MainEntity.Type = "Question"
	jsonLD.MainEntity.Name = detail.Title
	jsonLD.MainEntity.Text = detail.HTML
	jsonLD.MainEntity.AnswerCount = int(answerCount)
	jsonLD.MainEntity.UpvoteCount = detail.VoteCount
	jsonLD.MainEntity.DateCreated = time.Unix(detail.CreateTime, 0)
	jsonLD.MainEntity.Author.Type = "Person"
	jsonLD.MainEntity.Author.Name = detail.UserInfo.DisplayName
	answerList := make([]*schema.SuggestedAnswerItem, 0)
	for _, answer := range answers {
		if answer.Accepted == schema.AnswerAcceptedEnable {
			acceptedAnswerItem := &schema.AcceptedAnswerItem{}
			acceptedAnswerItem.Type = "Answer"
			acceptedAnswerItem.Text = answer.HTML
			acceptedAnswerItem.DateCreated = time.Unix(answer.CreateTime, 0)
			acceptedAnswerItem.UpvoteCount = answer.VoteCount
			acceptedAnswerItem.URL = fmt.Sprintf("%s/%s", site.Canonical, answer.ID)
			acceptedAnswerItem.Author.Type = "Person"
			acceptedAnswerItem.Author.Name = answer.UserInfo.DisplayName
			jsonLD.MainEntity.AcceptedAnswer = acceptedAnswerItem
		} else {
			item := &schema.SuggestedAnswerItem{}
			item.Type = "Answer"
			item.Text = answer.HTML
			item.DateCreated = time.Unix(answer.CreateTime, 0)
			item.UpvoteCount = answer.VoteCount
			item.URL = fmt.Sprintf("%s/%s", site.Canonical, answer.ID)
			item.Author.Type = "Person"
			item.Author.Name = answer.UserInfo.DisplayName
			answerList = append(answerList, item)
		}

	}
	jsonLD.MainEntity.SuggestedAnswer = answerList
	jsonLDStr, err := json.Marshal(jsonLD)
	if err == nil {
		site.JsonLD = `<script data-react-helmet="true" type="application/ld+json">` + string(jsonLDStr) + ` </script>`
	}

	site.Description = htmltext.FetchExcerpt(detail.HTML, "...", 240)
	tags := make([]string, 0)
	for _, tag := range detail.Tags {
		tags = append(tags, tag.DisplayName)
	}
	site.Keywords = strings.Replace(strings.Trim(fmt.Sprint(tags), "[]"), " ", ",", -1)
	site.Title = fmt.Sprintf("%s - %s", detail.Title, site.General.Name)
	tc.html(ctx, http.StatusOK, "question-detail.html", site, gin.H{
		"id":       id,
		"answerid": answerid,
		"detail":   detail,
		"answers":  answers,
		"comments": comments,
	})
}

// TagList tags list
func (tc *TemplateController) TagList(ctx *gin.Context) {
	req := &schema.GetTagWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	data, err := tc.templateRenderController.TagList(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	page := templaterender.Paginator(req.Page, req.PageSize, data.Count)

	site := tc.SiteInfo(ctx)
	site.Canonical = fmt.Sprintf("%s/tags", site.General.SiteUrl)
	if req.Page > 1 {
		site.Canonical = fmt.Sprintf("%s/tags?page=%d", site.General.SiteUrl, req.Page)
	}
	site.Title = fmt.Sprintf("%s - %s", "Tags", site.General.Name)
	tc.html(ctx, http.StatusOK, "tags.html", site, gin.H{
		"page": page,
		"data": data,
	})
}

// TagInfo taginfo
func (tc *TemplateController) TagInfo(ctx *gin.Context) {
	tag := ctx.Param("tag")
	req := &schema.GetTamplateTagInfoReq{}
	if handler.BindAndCheck(ctx, req) {
		tc.Page404(ctx)
		return
	}
	nowPage := req.Page
	req.Name = tag
	taginifo, questionList, questionCount, err := tc.templateRenderController.TagInfo(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	page := templaterender.Paginator(nowPage, req.PageSize, questionCount)

	site := tc.SiteInfo(ctx)
	site.Canonical = fmt.Sprintf("%s/tags/%s", site.General.SiteUrl, tag)
	if req.Page > 1 {
		site.Canonical = fmt.Sprintf("%s/tags/%s?page=%d", site.General.SiteUrl, tag, req.Page)
	}
	site.Description = htmltext.FetchExcerpt(taginifo.ParsedText, "...", 240)
	if len(taginifo.ParsedText) == 0 {
		site.Description = "The tag has no description."
	}
	site.Keywords = taginifo.DisplayName

	UrlUseTitle := false
	if site.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitle {
		UrlUseTitle = true
	}
	site.Title = fmt.Sprintf("'%s' Questions - %s", taginifo.DisplayName, site.General.Name)
	tc.html(ctx, http.StatusOK, "tag-detail.html", site, gin.H{
		"tag":           taginifo,
		"questionList":  questionList,
		"questionCount": questionCount,
		"useTitle":      UrlUseTitle,
		"page":          page,
	})
}

// UserInfo user info
func (tc *TemplateController) UserInfo(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		tc.Page404(ctx)
		return
	}

	exist := checker.IsUsersIgnorePath(username)
	if exist {
		//file, err := ui.Build.ReadFile("build/index.html")
		//if err != nil {
		//	log.Error(err)
		//	tc.Page404(ctx)
		//	return
		//}
		ctx.Header("content-type", "text/html;charset=utf-8")
		//ctx.String(http.StatusOK, string(file))
		return
	}
	req := &schema.GetOtherUserInfoByUsernameReq{}
	req.Username = username
	userinfo, err := tc.templateRenderController.UserInfo(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}

	site := tc.SiteInfo(ctx)
	site.Canonical = fmt.Sprintf("%s/users/%s", site.General.SiteUrl, username)
	site.Title = fmt.Sprintf("%s - %s", username, site.General.Name)
	tc.html(ctx, http.StatusOK, "homepage.html", site, gin.H{
		"userinfo": userinfo,
		"bio":      template.HTML(userinfo.BioHTML),
	})

}

func (tc *TemplateController) Page404(ctx *gin.Context) {
	tc.html(ctx, http.StatusNotFound, "404.html", tc.SiteInfo(ctx), gin.H{})
}

func (tc *TemplateController) html(ctx *gin.Context, code int, tpl string, site *schema.TemplateSiteInfoResp, data gin.H) {
	data["siteinfo"] = site
	data["scriptPath"] = tc.scriptPath
	data["cssPath"] = tc.cssPath
	data["keywords"] = site.Keywords
	if site.Description == "" {
		site.Description = site.General.Description
	}
	data["title"] = site.Title
	if site.Title == "" {
		data["title"] = site.General.Name
	}
	data["description"] = site.Description
	data["language"] = utils.GetLang(ctx)
	data["timezone"] = site.Interface.TimeZone
	language := strings.Replace(site.Interface.Language, "_", "-", -1)
	data["lang"] = language
	data["HeadCode"] = site.CustomCssHtml.CustomHead
	data["HeaderCode"] = site.CustomCssHtml.CustomHeader
	data["FooterCode"] = site.CustomCssHtml.CustomFooter
	data["Version"] = constant.Version
	data["Revision"] = constant.Revision
	_, ok := data["path"]
	if !ok {
		data["path"] = ""
	}
	ctx.Header("X-Frame-Options", "DENY")
	ctx.HTML(code, tpl, data)
}

func (tc *TemplateController) Sitemap(ctx *gin.Context) {
	if tc.checkPrivateMode(ctx) {
		tc.Page404(ctx)
		return
	}
	tc.templateRenderController.Sitemap(ctx)
}

func (tc *TemplateController) SitemapPage(ctx *gin.Context) {
	if tc.checkPrivateMode(ctx) {
		tc.Page404(ctx)
		return
	}
	page := 0
	pageParam := ctx.Param("page")
	pageRegexp := regexp.MustCompile(`question-(.*).xml`)
	pageStr := pageRegexp.FindStringSubmatch(pageParam)
	if len(pageStr) != 2 {
		tc.Page404(ctx)
		return
	}
	page = converter.StringToInt(pageStr[1])
	if page == 0 {
		tc.Page404(ctx)
		return
	}
	err := tc.templateRenderController.SitemapPage(ctx, page)
	if err != nil {
		tc.Page404(ctx)
		return
	}
}

func (tc *TemplateController) checkPrivateMode(ctx *gin.Context) bool {
	resp, err := services.SiteInfoService.GetSiteLogin(ctx)
	if err != nil {
		log.Error(err)
		return false
	}
	if resp.LoginRequired {
		return true
	}
	return false
}
*/
