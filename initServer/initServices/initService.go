package services

import (
	"github.com/lawyer/service"
	"github.com/lawyer/service/action"
	activity2 "github.com/lawyer/service/activity"
	activity_common2 "github.com/lawyer/service/activity_common"
	"github.com/lawyer/service/activity_queue"
	answercommon "github.com/lawyer/service/answer_common"
	auth2 "github.com/lawyer/service/auth"
	collectioncommon "github.com/lawyer/service/collection_common"
	comment2 "github.com/lawyer/service/comment"
	"github.com/lawyer/service/comment_common"
	"github.com/lawyer/service/dashboard"
	export2 "github.com/lawyer/service/export"
	"github.com/lawyer/service/follow"
	meta2 "github.com/lawyer/service/meta"
	"github.com/lawyer/service/notice_queue"
	"github.com/lawyer/service/notification"
	notificationcommon "github.com/lawyer/service/notification_common"
	"github.com/lawyer/service/object_info"
	"github.com/lawyer/service/plugin_common"
	questioncommon "github.com/lawyer/service/question_common"
	rank2 "github.com/lawyer/service/rank"
	reason2 "github.com/lawyer/service/reason"
	report2 "github.com/lawyer/service/report"
	"github.com/lawyer/service/report_admin"
	"github.com/lawyer/service/report_handle_admin"
	"github.com/lawyer/service/revision_common"
	role2 "github.com/lawyer/service/role"
	"github.com/lawyer/service/search_parser"
	"github.com/lawyer/service/siteinfo"
	"github.com/lawyer/service/siteinfo_common"
	tag2 "github.com/lawyer/service/tag"
	tag_common2 "github.com/lawyer/service/tag_common"
	"github.com/lawyer/service/uploader"
	"github.com/lawyer/service/user_admin"
	usercommon "github.com/lawyer/service/user_common"
	user_external_login2 "github.com/lawyer/service/user_external_login"
	user_notification_config2 "github.com/lawyer/service/user_notification_config"
)

var (
	SiteInfoCommonService            siteinfo_common.SiteInfoCommonService
	AuthService                      *auth2.AuthService
	EmailService                     *export2.EmailService
	RoleService                      *role2.RoleService
	UserRoleRelService               *role2.UserRoleRelService
	UserCommon                       *usercommon.UserCommon
	UserNotificationConfigService    *user_notification_config2.UserNotificationConfigService
	UserExternalLoginService         *user_external_login2.UserExternalLoginService
	UserService                      *service.UserService
	CaptchaService                   *action.CaptchaService
	RevisionService                  *revision_common.RevisionService
	ActivityQueueService             activity_queue.ActivityQueueService
	TagCommonService                 *tag_common2.TagCommonService
	ObjService                       *object_info.ObjService
	NotificationQueueService         notice_queue.NotificationQueueService
	ExternalNotificationQueueService notice_queue.ExternalNotificationQueueService
	CommentService                   *comment2.CommentService

	RolePowerRelService *role2.RolePowerRelService
	RankService         *rank2.RankService

	ReportService *report2.ReportService
	VoteService   *service.VoteService
	TagService    *tag2.TagService
	FollowService *follow.FollowService

	CollectionCommon            *collectioncommon.CollectionCommon
	AnswerCommon                *answercommon.AnswerCommon
	MetaService                 *meta2.MetaService
	QuestionCommon              *questioncommon.QuestionCommon
	CollectionService           *service.CollectionService
	ExternalNotificationService *notification.ExternalNotificationService
	AnswerActivityService       *activity2.AnswerActivityService
	QuestionService             *service.QuestionService
	AnswerService               *service.AnswerService
	SearchParser                *search_parser.SearchParser
	SearchService               *service.SearchService
	ServiceRevisionService      *service.RevisionService
	ReportHandle                *report_handle_admin.ReportHandle
	ReportAdminService          *report_admin.ReportAdminService
	UserAdminService            *user_admin.UserAdminService
	ReasonService               *reason2.ReasonService
	SiteInfoService             *siteinfo.SiteInfoService
	NotificationCommon          *notificationcommon.NotificationCommon
	NotificationService         *notification.NotificationService
	ActivityCommon              *activity_common2.ActivityCommon
	CommentCommonService        *comment_common.CommentCommonService
	PluginCommonService         *plugin_common.PluginCommonService
	UploaderService             uploader.UploaderService
	DashboardService            dashboard.DashboardService
	ActivityService             *activity2.ActivityService
)

func InitServices() {

	SiteInfoCommonService = siteinfo_common.NewSiteInfoCommonService()
	AuthService = auth2.NewAuthService()
	EmailService = export2.NewEmailService()
	RoleService = role2.NewRoleService()
	UserRoleRelService = role2.NewUserRoleRelService()
	UserCommon = usercommon.NewUserCommon()
	UserNotificationConfigService = user_notification_config2.NewUserNotificationConfigService()
	UserExternalLoginService = user_external_login2.NewUserExternalLoginService()
	UserService = service.NewUserService()
	CaptchaService = action.NewCaptchaService()
	RevisionService = revision_common.NewRevisionService()
	ActivityQueueService = activity_queue.NewActivityQueueService()
	TagCommonService = tag_common2.NewTagCommonService()
	ObjService = object_info.NewObjService()

	NotificationQueueService = notice_queue.NewNotificationQueueService()
	ExternalNotificationQueueService = notice_queue.NewNewQuestionNotificationQueueService()
	CommentService = comment2.NewCommentService()

	RolePowerRelService = role2.NewRolePowerRelService()
	RankService = rank2.NewRankService()

	ReportService = report2.NewReportService()
	VoteService = service.NewVoteService()
	TagService = tag2.NewTagService()
	FollowService = follow.NewFollowService()

	CollectionCommon = collectioncommon.NewCollectionCommon()
	AnswerCommon = answercommon.NewAnswerCommon()
	MetaService = meta2.NewMetaService()
	QuestionCommon = questioncommon.NewQuestionCommon()
	CollectionService = service.NewCollectionService()
	ExternalNotificationService = notification.NewExternalNotificationService()
	AnswerActivityService = activity2.NewAnswerActivityService()
	QuestionService = service.NewQuestionService()
	AnswerService = service.NewAnswerService()
	SearchParser = search_parser.NewSearchParser()
	SearchService = service.NewSearchService()
	ServiceRevisionService = service.NewRevisionService()
	ReportHandle = report_handle_admin.NewReportHandle()
	ReportAdminService = report_admin.NewReportAdminService()
	UserAdminService = user_admin.NewUserAdminService()
	ReasonService = reason2.NewReasonService()
	SiteInfoService = siteinfo.NewSiteInfoService()
	NotificationCommon = notificationcommon.NewNotificationCommon()
	NotificationService = notification.NewNotificationService()
	ActivityCommon = activity_common2.NewActivityCommon()
	CommentCommonService = comment_common.NewCommentCommonService()
	PluginCommonService = plugin_common.NewPluginCommonService()
	UploaderService = uploader.NewUploaderService()
	DashboardService = dashboard.NewDashboardService()
	ActivityService = activity2.NewActivityService()
}
