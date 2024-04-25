package services

import (
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/config"
	"github.com/lawyer/service"
	"github.com/lawyer/service/action"
	"github.com/lawyer/service/activity"
	"github.com/lawyer/service/activity_common"
	"github.com/lawyer/service/activity_queue"
	answercommon "github.com/lawyer/service/answer_common"
	"github.com/lawyer/service/auth"
	collectioncommon "github.com/lawyer/service/collection_common"
	"github.com/lawyer/service/comment"
	"github.com/lawyer/service/comment_common"
	"github.com/lawyer/service/dashboard"
	"github.com/lawyer/service/export"
	"github.com/lawyer/service/follow"
	"github.com/lawyer/service/meta"
	"github.com/lawyer/service/notice_queue"
	"github.com/lawyer/service/notification"
	notificationcom "github.com/lawyer/service/notification_common"
	"github.com/lawyer/service/object_info"
	"github.com/lawyer/service/plugin_common"
	questioncommon "github.com/lawyer/service/question_common"
	"github.com/lawyer/service/rank"
	"github.com/lawyer/service/reason"
	"github.com/lawyer/service/report"
	"github.com/lawyer/service/report_admin"
	"github.com/lawyer/service/report_handle_admin"
	"github.com/lawyer/service/revision_common"
	"github.com/lawyer/service/role"
	"github.com/lawyer/service/search_parser"
	"github.com/lawyer/service/siteinfo"
	"github.com/lawyer/service/siteinfo_common"
	"github.com/lawyer/service/tag"
	"github.com/lawyer/service/tag_common"
	"github.com/lawyer/service/uploader"
	"github.com/lawyer/service/user_admin"
	usercommon "github.com/lawyer/service/user_common"
	"github.com/lawyer/service/user_external_login"
	"github.com/lawyer/service/user_notification_config"
	"github.com/segmentfault/pacman/i18n"
)

var (
	SiteInfoCommonService            siteinfo_common.SiteInfoCommonService
	AuthService                      *auth.AuthService
	EmailService                     *export.EmailService
	RoleService                      *role.RoleService
	UserRoleRelService               *role.UserRoleRelService
	UserCommon                       *usercommon.UserCommon
	UserNotificationConfigService    *user_notification_config.UserNotificationConfigService
	UserExternalLoginService         *user_external_login.UserExternalLoginService
	UserService                      *service.UserService
	CaptchaService                   *action.CaptchaService
	RevisionService                  *revision_common.RevisionService
	ActivityQueueService             activity_queue.ActivityQueueService
	TagCommonService                 *tag_common.TagCommonService
	ObjService                       *object_info.ObjService
	NotificationQueueService         notice_queue.NotificationQueueService
	ExternalNotificationQueueService notice_queue.ExternalNotificationQueueService
	CommentService                   *comment.CommentService

	RolePowerRelService *role.RolePowerRelService
	RankService         *rank.RankService

	ReportService *report.ReportService
	VoteService   *service.VoteService
	TagService    *tag.TagService
	FollowService *follow.FollowService

	CollectionCommon            *collectioncommon.CollectionCommon
	AnswerCommon                *answercommon.AnswerCommon
	MetaService                 *meta.MetaService
	QuestionCommon              *questioncommon.QuestionCommon
	CollectionService           *service.CollectionService
	ExternalNotificationService *notification.ExternalNotificationService
	AnswerActivityService       *activity.AnswerActivityService
	QuestionService             *service.QuestionService
	AnswerService               *service.AnswerService
	SearchParser                *search_parser.SearchParser
	SearchService               *service.SearchService
	ServiceRevisionService      *service.RevisionService
	ReportHandle                *report_handle_admin.ReportHandle
	ReportAdminService          *report_admin.ReportAdminService
	UserAdminService            *user_admin.UserAdminService
	ReasonService               *reason.ReasonService
	SiteInfoService             *siteinfo.SiteInfoService
	NotificationCommon          *notificationcom.NotificationCommon
	NotificationService         *notification.NotificationService
	ActivityCommon              *activity_common.ActivityCommon
	CommentCommonService        *comment_common.CommentCommonService
	PluginCommonService         *plugin_common.PluginCommonService
	UploaderService             uploader.UploaderService
	DashboardService            dashboard.DashboardService
	ActivityService             *activity.ActivityService
)

var (
	I18nTranslator i18n.Translator
)

// todo
func InitTranslator(i18nConf *config.I18n) (err error) {
	I18nTranslator, err = translator.NewTranslator(i18nConf)
	return err
}

func InitServices() {

	SiteInfoCommonService = siteinfo_common.NewSiteInfoCommonService()
	AuthService = auth.NewAuthService()
	EmailService = export.NewEmailService()
	RoleService = role.NewRoleService()
	UserRoleRelService = role.NewUserRoleRelService()
	UserCommon = usercommon.NewUserCommon()
	UserNotificationConfigService = user_notification_config.NewUserNotificationConfigService()
	UserExternalLoginService = user_external_login.NewUserExternalLoginService()
	UserService = service.NewUserService()
	CaptchaService = action.NewCaptchaService()
	RevisionService = revision_common.NewRevisionService()
	ActivityQueueService = activity_queue.NewActivityQueueService()
	TagCommonService = tag_common.NewTagCommonService()
	ObjService = object_info.NewObjService()

	NotificationQueueService = notice_queue.NewNotificationQueueService()
	ExternalNotificationQueueService = notice_queue.NewNewQuestionNotificationQueueService()
	CommentService = comment.NewCommentService()

	RolePowerRelService = role.NewRolePowerRelService()
	RankService = rank.NewRankService()

	ReportService = report.NewReportService()
	VoteService = service.NewVoteService()
	TagService = tag.NewTagService()
	FollowService = follow.NewFollowService()

	CollectionCommon = collectioncommon.NewCollectionCommon()
	AnswerCommon = answercommon.NewAnswerCommon()
	MetaService = meta.NewMetaService()
	QuestionCommon = questioncommon.NewQuestionCommon()
	CollectionService = service.NewCollectionService()
	ExternalNotificationService = notification.NewExternalNotificationService()
	AnswerActivityService = activity.NewAnswerActivityService()
	QuestionService = service.NewQuestionService()
	AnswerService = service.NewAnswerService()
	SearchParser = search_parser.NewSearchParser()
	SearchService = service.NewSearchService()
	ServiceRevisionService = service.NewRevisionService()
	ReportHandle = report_handle_admin.NewReportHandle()
	ReportAdminService = report_admin.NewReportAdminService()
	UserAdminService = user_admin.NewUserAdminService()
	ReasonService = reason.NewReasonService()
	SiteInfoService = siteinfo.NewSiteInfoService()
	NotificationCommon = notificationcom.NewNotificationCommon()
	NotificationService = notification.NewNotificationService()
	ActivityCommon = activity_common.NewActivityCommon()
	CommentCommonService = comment_common.NewCommentCommonService()
	PluginCommonService = plugin_common.NewPluginCommonService()
	UploaderService = uploader.NewUploaderService()
	DashboardService = dashboard.NewDashboardService()
	ActivityService = activity.NewActivityService()
}
