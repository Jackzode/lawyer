package export

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/base/translator"
	c "github.com/lawyer/commons/config"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/utils"
	services "github.com/lawyer/initServer/initServices"
	"github.com/lawyer/pkg/display"
	"github.com/lawyer/repo"
	"github.com/lawyer/service/config"
	"mime"
	"os"
	"strings"
	"time"

	"github.com/lawyer/commons/schema"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"golang.org/x/net/context"
	"gopkg.in/gomail.v2"
)

// EmailService kit service
type EmailService struct {
	configService *config.ConfigService
}

// EmailRepo email repository
type EmailRepo interface {
	SetCode(ctx context.Context, code, content string, duration time.Duration) error
	VerifyCode(ctx context.Context, code string) (content string, err error)
}

// NewEmailService email service
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SaveCode save code
func (es *EmailService) SaveCode(ctx context.Context, code, codeContent string) {
	err := repo.EmailRepo.SetCode(ctx, code, codeContent, 10*time.Minute)
	if err != nil {
		log.Error(err)
	}
}

// SendAndSaveCode send email and save code
func (es *EmailService) SendAndSaveCode(ctx context.Context, toEmailAddr, subject, body, code, codeContent string) {
	es.Send(ctx, toEmailAddr, subject, body)
	err := repo.EmailRepo.SetCode(ctx, code, codeContent, 10*time.Minute)
	if err != nil {
		log.Error(err)
	}
}

// SendAndSaveCodeWithTime send email and save code
func (es *EmailService) SendAndSaveCodeWithTime(
	ctx context.Context, toEmailAddr, subject, body, code, codeContent string, duration time.Duration) {
	es.Send(ctx, toEmailAddr, subject, body)
	err := repo.EmailRepo.SetCode(ctx, code, codeContent, duration)
	if err != nil {
		log.Error(err)
	}
}

// Send email send
func (es *EmailService) Send(ctx context.Context, toEmailAddr, subject, body string) {
	log.Infof("try to send email to %s", toEmailAddr)
	ec, err := es.GetEmailConfig(ctx)
	if err != nil {
		log.Errorf("get email config failed: %s", err)
		return
	}
	if len(ec.SMTPHost) == 0 {
		log.Warnf("smtp host is empty, skip send email")
		return
	}

	m := gomail.NewMessage()
	fromName := mime.QEncoding.Encode("utf-8", ec.FromName)
	m.SetHeader("From", fmt.Sprintf("%s <%s>", fromName, ec.FromEmail))
	m.SetHeader("To", toEmailAddr)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(ec.SMTPHost, ec.SMTPPort, ec.SMTPUsername, ec.SMTPPassword)
	if ec.IsSSL() {
		d.SSL = true
	}
	if len(os.Getenv("SKIP_SMTP_TLS_VERIFY")) > 0 {
		d.TLSConfig = &tls.Config{ServerName: d.Host, InsecureSkipVerify: true}
	}
	if err := d.DialAndSend(m); err != nil {
		log.Errorf("send email to %s failed: %s", toEmailAddr, err)
	} else {
		log.Infof("send email to %s success", toEmailAddr)
	}
}

// VerifyUrlExpired 根据code从缓存中获取content
func (es *EmailService) VerifyUrlExpired(ctx context.Context, code string) (content string) {
	content, err := repo.EmailRepo.VerifyCode(ctx, code)
	if err != nil {
		log.Error(err)
	}
	return content
}

func (es *EmailService) RegisterTemplate(ctx context.Context, registerUrl string) (title, body string, err error) {
	siteInfo, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return
	}
	templateData := &schema.RegisterTemplateData{
		SiteName:    siteInfo.Name,
		RegisterUrl: registerUrl,
	}

	lang := utils.GetLangByCtx(ctx)
	title = translator.TrWithData(lang, constant.EmailTplKeyRegisterTitle, templateData)
	body = translator.TrWithData(lang, constant.EmailTplKeyRegisterBody, templateData)
	return title, body, nil
}

func (es *EmailService) PassResetTemplate(ctx context.Context, passResetUrl string) (title, body string, err error) {
	siteInfo, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return
	}

	templateData := &schema.PassResetTemplateData{SiteName: siteInfo.Name, PassResetUrl: passResetUrl}

	lang := utils.GetLangByCtx(ctx)
	title = translator.TrWithData(lang, constant.EmailTplKeyPassResetTitle, templateData)
	body = translator.TrWithData(lang, constant.EmailTplKeyPassResetBody, templateData)
	return title, body, nil
}

func (es *EmailService) ChangeEmailTemplate(ctx context.Context, changeEmailUrl string) (title, body string, err error) {
	siteInfo, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return
	}
	templateData := &schema.ChangeEmailTemplateData{
		SiteName:       siteInfo.Name,
		ChangeEmailUrl: changeEmailUrl,
	}

	lang := utils.GetLangByCtx(ctx)
	title = translator.TrWithData(lang, constant.EmailTplKeyChangeEmailTitle, templateData)
	body = translator.TrWithData(lang, constant.EmailTplKeyChangeEmailBody, templateData)
	return title, body, nil
}

// TestTemplate send test email template parse
func (es *EmailService) TestTemplate(ctx context.Context) (title, body string, err error) {
	siteInfo, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return
	}
	templateData := &schema.TestTemplateData{SiteName: siteInfo.Name}

	lang := utils.GetLangByCtx(ctx)
	title = translator.TrWithData(lang, constant.EmailTplKeyTestTitle, templateData)
	body = translator.TrWithData(lang, constant.EmailTplKeyTestBody, templateData)
	return title, body, nil
}

// NewAnswerTemplate new answer template
func (es *EmailService) NewAnswerTemplate(ctx context.Context, raw *schema.NewAnswerTemplateRawData) (
	title, body string, err error) {
	siteInfo, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return
	}
	seoInfo, err := services.SiteInfoCommonService.GetSiteSeo(ctx)
	if err != nil {
		return
	}
	templateData := &schema.NewAnswerTemplateData{
		SiteName:       siteInfo.Name,
		DisplayName:    raw.AnswerUserDisplayName,
		QuestionTitle:  raw.QuestionTitle,
		AnswerUrl:      display.AnswerURL(seoInfo.Permalink, siteInfo.SiteUrl, raw.QuestionID, raw.QuestionTitle, raw.AnswerID),
		AnswerSummary:  raw.AnswerSummary,
		UnsubscribeUrl: fmt.Sprintf("%s/users/unsubscribe?code=%s", siteInfo.SiteUrl, raw.UnsubscribeCode),
	}

	lang := utils.GetLangByCtx(ctx)
	title = translator.TrWithData(lang, constant.EmailTplKeyNewAnswerTitle, templateData)
	body = translator.TrWithData(lang, constant.EmailTplKeyNewAnswerBody, templateData)
	return title, body, nil
}

// NewInviteAnswerTemplate new invite answer template
func (es *EmailService) NewInviteAnswerTemplate(ctx context.Context, raw *schema.NewInviteAnswerTemplateRawData) (
	title, body string, err error) {
	siteInfo, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return
	}
	seoInfo, err := services.SiteInfoCommonService.GetSiteSeo(ctx)
	if err != nil {
		return
	}
	templateData := &schema.NewInviteAnswerTemplateData{
		SiteName:       siteInfo.Name,
		DisplayName:    raw.InviterDisplayName,
		QuestionTitle:  raw.QuestionTitle,
		InviteUrl:      display.QuestionURL(seoInfo.Permalink, siteInfo.SiteUrl, raw.QuestionID, raw.QuestionTitle),
		UnsubscribeUrl: fmt.Sprintf("%s/users/unsubscribe?code=%s", siteInfo.SiteUrl, raw.UnsubscribeCode),
	}

	lang := utils.GetLangByCtx(ctx)
	title = translator.TrWithData(lang, constant.EmailTplKeyInvitedAnswerTitle, templateData)
	body = translator.TrWithData(lang, constant.EmailTplKeyInvitedAnswerBody, templateData)
	return title, body, nil
}

// NewCommentTemplate new comment template
func (es *EmailService) NewCommentTemplate(ctx context.Context, raw *schema.NewCommentTemplateRawData) (
	title, body string, err error) {
	siteInfo, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return
	}
	seoInfo, err := services.SiteInfoCommonService.GetSiteSeo(ctx)
	if err != nil {
		return
	}
	templateData := &schema.NewCommentTemplateData{
		SiteName:       siteInfo.Name,
		DisplayName:    raw.CommentUserDisplayName,
		QuestionTitle:  raw.QuestionTitle,
		CommentSummary: raw.CommentSummary,
		UnsubscribeUrl: fmt.Sprintf("%s/users/unsubscribe?code=%s", siteInfo.SiteUrl, raw.UnsubscribeCode),
	}
	templateData.CommentUrl = display.CommentURL(seoInfo.Permalink,
		siteInfo.SiteUrl, raw.QuestionID, raw.QuestionTitle, raw.AnswerID, raw.CommentID)

	lang := utils.GetLangByCtx(ctx)
	title = translator.TrWithData(lang, constant.EmailTplKeyNewCommentTitle, templateData)
	body = translator.TrWithData(lang, constant.EmailTplKeyNewCommentBody, templateData)
	return title, body, nil
}

// NewQuestionTemplate new question template
func (es *EmailService) NewQuestionTemplate(ctx context.Context, raw *schema.NewQuestionTemplateRawData) (
	title, body string, err error) {
	siteInfo, err := services.SiteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return
	}
	seoInfo, err := services.SiteInfoCommonService.GetSiteSeo(ctx)
	if err != nil {
		return
	}
	templateData := &schema.NewQuestionTemplateData{
		SiteName:       siteInfo.Name,
		QuestionTitle:  raw.QuestionTitle,
		Tags:           strings.Join(raw.Tags, ", "),
		UnsubscribeUrl: fmt.Sprintf("%s/users/unsubscribe?code=%s", siteInfo.SiteUrl, raw.UnsubscribeCode),
	}
	templateData.QuestionUrl = display.QuestionURL(
		seoInfo.Permalink, siteInfo.SiteUrl, raw.QuestionID, raw.QuestionTitle)

	lang := utils.GetLangByCtx(ctx)
	title = translator.TrWithData(lang, constant.EmailTplKeyNewQuestionTitle, templateData)
	body = translator.TrWithData(lang, constant.EmailTplKeyNewQuestionBody, templateData)
	return title, body, nil
}

func (es *EmailService) GetEmailConfig(ctx context.Context) (ec *c.EmailConfig, err error) {
	emailConf, err := utils.GetStringValue(ctx, "email.config")
	if err != nil {
		return nil, err
	}
	ec = &c.EmailConfig{}
	err = json.Unmarshal([]byte(emailConf), ec)
	if err != nil {
		log.Errorf("old email config format is invalid, you need to update smtp config: %v", err)
		return nil, errors.BadRequest(reason.SiteInfoConfigNotFound)
	}
	return ec, nil
}

// SetEmailConfig set email config
func (es *EmailService) SetEmailConfig(ctx context.Context, ec *c.EmailConfig) (err error) {
	data, _ := json.Marshal(ec)
	return es.configService.UpdateConfig(ctx, "email.config", string(data))
}
