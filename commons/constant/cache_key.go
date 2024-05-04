package constant

import "time"

const (
	UserRegisterInfoKey                        = "lawyer:user:register:"
	UserRegisterInfoTime                       = 10 * time.Minute
	UserCacheInfoKey                           = "lawyer:user:cache:"
	UserCacheInfoChangedCacheTime              = 7 * 24 * time.Hour
	UserTokenCacheKey                          = "lawyer:user:token:"
	UserTokenCacheTime                         = 7 * 24 * time.Hour
	UserVisitTokenCacheKey                     = "lawyer:user:visit:"
	UserVisitCacheTime                         = 7 * 24 * 60 * 60
	UserVisitCookiesCacheKey                   = "visit"
	AdminTokenCacheKey                         = "lawyer:admin:token:"
	AdminTokenCacheTime                        = 7 * 24 * time.Hour
	UserTokenMappingCacheKey                   = "lawyer:user-token:mapping:"
	SiteInfoCacheKey                           = "lawyer:site-info:"
	SiteInfoCacheTime                          = 1 * time.Hour
	ConfigID2KEYCacheKeyPrefix                 = "lawyer:config:id:"
	ConfigKEY2ContentCacheKeyPrefix            = "lawyer:config:key:"
	ConfigCacheTime                            = 1 * time.Hour
	ConnectorUserExternalInfoCacheKey          = "lawyer:connector:"
	ConnectorUserExternalInfoCacheTime         = 10 * time.Minute
	SiteMapQuestionCacheKeyPrefix              = "lawyer:sitemap:question:%d"
	SiteMapQuestionCacheTime                   = time.Hour
	SitemapMaxSize                             = 50000
	NewQuestionNotificationLimitCacheKeyPrefix = "lawyer:new-question-notification-limit:"
	NewQuestionNotificationLimitCacheTime      = 7 * 24 * time.Hour
	NewQuestionNotificationLimitMax            = 50
	RateLimitCacheKeyPrefix                    = "lawyer:rate-limit:"
	RateLimitCacheTime                         = 5 * time.Minute
)
