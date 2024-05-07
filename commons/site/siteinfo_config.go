package site

import (
	"fmt"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/schema"
	"github.com/spf13/viper"
)

var Config *siteInfoConfig

type siteInfoConfig struct {
	ContactEmail            string `mapstructure:"contact_email" `
	Language                string `mapstructure:"language" `
	TimeZone                string `mapstructure:"time_zone"`
	SiteName                string `mapstructure:"site_name"`
	SiteUrl                 string `mapstructure:"site_url"`
	DefaultAvatar           string `mapstructure:"default_avatar"`
	GravatarBaseUrl         string `mapstructure:"gravatar_base_url"`
	AllowEmailRegistrations bool   `mapstructure:"allow_email_registrations"`
	AllowNewRegistrations   bool   `mapstructure:"allow_new_registrations"`
	AllowPasswordLogin      bool   `mapstructure:"allow_password_login"`
	LoginRequired           bool   `mapstructure:"login_required"`
	AllowUpdateAvatar       bool   `mapstructure:"allow_update_avatar"`
	AllowUpdateBio          bool   `mapstructure:"allow_update_bio"`
	AllowUpdateDisplayName  bool   `mapstructure:"allow_update_display_name"`
	AllowUpdateLocation     bool   `mapstructure:"allow_update_location"`
	AllowUpdateUsername     bool   `mapstructure:"allow_update_username"`
	AllowUpdateWebsite      bool   `mapstructure:"allow_update_website"`
	RestrictAnswer          bool   `mapstructure:"restrict_answer"`
	Level                   int    `mapstructure:"level"`
	Permalink               int    `mapstructure:"permalink"`
	Robots                  string `mapstructure:"robots"`
}

// NewSiteInfoCommonService new site info common service
func InitSiteInfo(config string) {
	v := viper.New()
	v.SetConfigFile(config)
	err := v.ReadInConfig()
	if err != nil {
		fmt.Println("read...", err.Error())
		glog.Slog.Error(err.Error())
		panic(err)
	}
	err = v.Unmarshal(&Config)
	if err != nil {
		fmt.Println("unmarshal..", err.Error())
		glog.Slog.Error(err.Error())
		panic(err)
	}
	return
}

// GetSiteGeneral get site info general
func (s *siteInfoConfig) GetSiteGeneral() (resp *schema.SiteGeneralResp) {
	resp = &schema.SiteGeneralResp{}
	resp.SiteUrl = s.SiteUrl
	resp.ContactEmail = s.ContactEmail
	resp.Description = ""
	resp.ShortDescription = ""
	resp.Name = s.SiteName
	return resp
}

// GetSiteInterface get site info interface
func (s *siteInfoConfig) GetSiteInterface() (resp *schema.SiteInterfaceResp) {
	resp = &schema.SiteInterfaceResp{}
	resp.TimeZone = s.TimeZone
	resp.Language = s.Language
	return resp
}

// GetSiteBranding get site info branding
func (s *siteInfoConfig) GetSiteBranding() (resp *schema.SiteBrandingResp) {
	resp = &schema.SiteBrandingResp{}
	resp.Logo = ""
	resp.Favicon = ""
	resp.MobileLogo = ""
	resp.SquareIcon = ""
	return resp
}

// GetSiteUsers get site info about users
func (s *siteInfoConfig) GetSiteUsers() (resp *schema.SiteUsersResp) {
	resp = &schema.SiteUsersResp{}
	resp.AllowUpdateAvatar = s.AllowUpdateAvatar
	resp.AllowUpdateDisplayName = s.AllowUpdateDisplayName
	resp.AllowUpdateBio = s.AllowUpdateBio
	resp.AllowUpdateLocation = s.AllowUpdateLocation
	resp.AllowUpdateWebsite = s.AllowUpdateWebsite
	resp.DefaultAvatar = s.DefaultAvatar
	resp.GravatarBaseURL = s.GravatarBaseUrl
	resp.AllowUpdateUsername = s.AllowUpdateUsername
	return resp
}

//func (s *siteInfoConfig) GetAvatarDefaultConfig() (string, string) {
//	gravatarBaseURL, defaultAvatar := constant.DefaultGravatarBaseURL, constant.DefaultAvatar
//	if len(s.GravatarBaseUrl) > 0 {
//		gravatarBaseURL = s.GravatarBaseUrl
//	}
//	if len(s.DefaultAvatar) > 0 {
//		defaultAvatar = s.DefaultAvatar
//	}
//	return gravatarBaseURL, defaultAvatar
//}

// GetSiteWrite get site info write
func (s *siteInfoConfig) GetSiteWrite() (resp *schema.SiteWriteResp) {
	resp = &schema.SiteWriteResp{}
	resp.RestrictAnswer = s.RestrictAnswer
	resp.UserID = ""
	resp.RecommendTags = []string{}
	resp.RequiredTag = false
	resp.RecommendTags = []string{}
	return resp
}

// GetSiteLegal get site info write
func (s *siteInfoConfig) GetSiteLegal() (resp *schema.SiteLegalResp) {
	resp = &schema.SiteLegalResp{}
	resp.PrivacyPolicyOriginalText = ""
	resp.TermsOfServiceOriginalText = ""
	resp.TermsOfServiceParsedText = ""
	resp.TermsOfServiceParsedText = ""
	return resp
}

// GetSiteLogin get site login config
func (s *siteInfoConfig) GetSiteLogin() (resp *schema.SiteLoginResp) {
	resp = &schema.SiteLoginResp{}
	resp.AllowPasswordLogin = s.AllowPasswordLogin
	resp.LoginRequired = s.LoginRequired
	resp.AllowEmailRegistrations = s.AllowEmailRegistrations
	resp.AllowNewRegistrations = s.AllowNewRegistrations
	return resp
}

// GetSiteCustomCssHTML get site custom css html config
func (s *siteInfoConfig) GetSiteCustomCssHTML() (resp *schema.SiteCustomCssHTMLResp) {
	resp = &schema.SiteCustomCssHTMLResp{}
	//if err = s.GetSiteInfoByType(ctx, constant.SiteTypeCustomCssHTML, resp); err != nil {
	//	return nil, err
	//}
	return resp
}

// GetSiteTheme get site theme
//func (s *siteInfoConfig) GetSiteTheme(ctx context.Context) (resp *schema.SiteThemeResp, err error) {
//
//	return resp, nil
//}

// GetSiteSeo get site seo
func (s *siteInfoConfig) GetSiteSeo() (resp *schema.SiteSeoResp) {
	resp = &schema.SiteSeoResp{}
	resp.Permalink = s.Permalink
	resp.Robots = s.Robots
	return resp
}

func (s *siteInfoConfig) EnableShortID() (enabled bool) {
	siteSeo := s.GetSiteSeo()
	return siteSeo.IsShortLink()
}
