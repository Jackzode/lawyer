package initServer

import (
	"github.com/apache/incubator-answer/initServer/data"
	"github.com/apache/incubator-answer/internal/service"
	"github.com/apache/incubator-answer/internal/service/action"
	activity2 "github.com/apache/incubator-answer/internal/service/activity"
	activity_common2 "github.com/apache/incubator-answer/internal/service/activity_common"
	"github.com/apache/incubator-answer/internal/service/activity_queue"
	answercommon "github.com/apache/incubator-answer/internal/service/answer_common"
	auth2 "github.com/apache/incubator-answer/internal/service/auth"
	collectioncommon "github.com/apache/incubator-answer/internal/service/collection_common"
	comment2 "github.com/apache/incubator-answer/internal/service/comment"
	"github.com/apache/incubator-answer/internal/service/comment_common"
	"github.com/apache/incubator-answer/internal/service/dashboard"
	export2 "github.com/apache/incubator-answer/internal/service/export"
	"github.com/apache/incubator-answer/internal/service/follow"
	meta2 "github.com/apache/incubator-answer/internal/service/meta"
	"github.com/apache/incubator-answer/internal/service/notice_queue"
	"github.com/apache/incubator-answer/internal/service/notification"
	notificationcommon "github.com/apache/incubator-answer/internal/service/notification_common"
	"github.com/apache/incubator-answer/internal/service/object_info"
	"github.com/apache/incubator-answer/internal/service/plugin_common"
	questioncommon "github.com/apache/incubator-answer/internal/service/question_common"
	rank2 "github.com/apache/incubator-answer/internal/service/rank"
	reason2 "github.com/apache/incubator-answer/internal/service/reason"
	report2 "github.com/apache/incubator-answer/internal/service/report"
	"github.com/apache/incubator-answer/internal/service/report_admin"
	"github.com/apache/incubator-answer/internal/service/report_handle_admin"
	"github.com/apache/incubator-answer/internal/service/revision_common"
	role2 "github.com/apache/incubator-answer/internal/service/role"
	"github.com/apache/incubator-answer/internal/service/search_parser"
	sc "github.com/apache/incubator-answer/internal/service/service_config"
	"github.com/apache/incubator-answer/internal/service/siteinfo"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	tag2 "github.com/apache/incubator-answer/internal/service/tag"
	tag_common2 "github.com/apache/incubator-answer/internal/service/tag_common"
	"github.com/apache/incubator-answer/internal/service/uploader"
	"github.com/apache/incubator-answer/internal/service/user_admin"
	usercommon "github.com/apache/incubator-answer/internal/service/user_common"
	user_external_login2 "github.com/apache/incubator-answer/internal/service/user_external_login"
	user_notification_config2 "github.com/apache/incubator-answer/internal/service/user_notification_config"
)

var (
	siteInfoCommonService            siteinfo_common.SiteInfoCommonService
	authService                      *auth2.AuthService
	emailService                     *export2.EmailService
	roleService                      *role2.RoleService
	userRoleRelService               *role2.UserRoleRelService
	userCommon                       *usercommon.UserCommon
	userNotificationConfigService    *user_notification_config2.UserNotificationConfigService
	userExternalLoginService         *user_external_login2.UserExternalLoginService
	userService                      *service.UserService
	captchaService                   *action.CaptchaService
	revisionService                  *revision_common.RevisionService
	activityQueueService             activity_queue.ActivityQueueService
	tagCommonService                 *tag_common2.TagCommonService
	objService                       *object_info.ObjService
	notificationQueueService         notice_queue.NotificationQueueService
	externalNotificationQueueService notice_queue.ExternalNotificationQueueService
	commentService                   *comment2.CommentService

	rolePowerRelService *role2.RolePowerRelService
	rankService         *rank2.RankService

	reportService *report2.ReportService
	voteService   *service.VoteService
	tagService    *tag2.TagService
	followService *follow.FollowService

	collectionCommon            *collectioncommon.CollectionCommon
	answerCommon                *answercommon.AnswerCommon
	metaService                 *meta2.MetaService
	questionCommon              *questioncommon.QuestionCommon
	collectionService           *service.CollectionService
	externalNotificationService *notification.ExternalNotificationService
	answerActivityService       *activity2.AnswerActivityService
	questionService             *service.QuestionService
	answerService               *service.AnswerService
	searchParser                *search_parser.SearchParser
	searchService               *service.SearchService
	serviceRevisionService      *service.RevisionService
	reportHandle                *report_handle_admin.ReportHandle
	reportAdminService          *report_admin.ReportAdminService
	userAdminService            *user_admin.UserAdminService
	reasonService               *reason2.ReasonService
	siteInfoService             *siteinfo.SiteInfoService
	notificationCommon          *notificationcommon.NotificationCommon
	notificationService         *notification.NotificationService
	activityCommon              *activity_common2.ActivityCommon
	commentCommonService        *comment_common.CommentCommonService
	pluginCommonService         *plugin_common.PluginCommonService
	uploaderService             uploader.UploaderService
	dashboardService            dashboard.DashboardService
	activityService             *activity2.ActivityService
)

func initModel(serviceConf *sc.ServiceConfig) {

	siteInfoCommonService = siteinfo_common.NewSiteInfoCommonService(siteInfoRepo)
	authService = auth2.NewAuthService(authRepo)
	emailService = export2.NewEmailService(emailRepo, siteInfoCommonService)
	roleService = role2.NewRoleService(roleRepo)
	userRoleRelService = role2.NewUserRoleRelService(userRoleRelRepo, roleService)
	userCommon = usercommon.NewUserCommon(userRepo, userRoleRelService, authService, siteInfoCommonService)
	userNotificationConfigService = user_notification_config2.NewUserNotificationConfigService(userRepo, userNotificationConfigRepo)
	userExternalLoginService = user_external_login2.NewUserExternalLoginService(userRepo, userCommon, userExternalLoginRepo, emailService, siteInfoCommonService, userActiveActivityRepo, userNotificationConfigService)
	userService = service.NewUserService(userRepo, userActiveActivityRepo, activityRepo, emailService, authService, siteInfoCommonService, userRoleRelService, userCommon, userExternalLoginService, userNotificationConfigRepo, userNotificationConfigService)
	captchaService = action.NewCaptchaService(captchaRepo)
	revisionService = revision_common.NewRevisionService(revisionRepo, userRepo)
	activityQueueService = activity_queue.NewActivityQueueService()
	tagCommonService = tag_common2.NewTagCommonService(tagCommonRepo, tagRelRepo, tagRepo, revisionService, siteInfoCommonService, activityQueueService)
	objService = object_info.NewObjService(answerRepo, questionRepo, commentCommonRepo, tagCommonRepo, tagCommonService)

	notificationQueueService = notice_queue.NewNotificationQueueService()
	externalNotificationQueueService = notice_queue.NewNewQuestionNotificationQueueService()
	commentService = comment2.NewCommentService(commentRepo, commentCommonRepo, userCommon, objService, voteRepo, emailService, userRepo, notificationQueueService, externalNotificationQueueService, activityQueueService)

	rolePowerRelService = role2.NewRolePowerRelService(rolePowerRelRepo, userRoleRelService)
	rankService = rank2.NewRankService(userCommon, userRankRepo, objService, userRoleRelService, rolePowerRelService)

	reportService = report2.NewReportService(reportRepo, objService)
	voteService = service.NewVoteService(serviceVoteRepo, questionRepo, answerRepo, commentCommonRepo, objService)
	tagService = tag2.NewTagService(tagRepo, tagCommonService, revisionService, followRepo, siteInfoCommonService, activityQueueService)
	followService = follow.NewFollowService(followFollowRepo, followRepo, tagCommonRepo)

	collectionCommon = collectioncommon.NewCollectionCommon(collectionRepo)
	answerCommon = answercommon.NewAnswerCommon(answerRepo)
	metaService = meta2.NewMetaService(metaRepo)
	questionCommon = questioncommon.NewQuestionCommon(questionRepo, answerRepo, voteRepo, followRepo, tagCommonService, userCommon, collectionCommon, answerCommon, metaService, activityQueueService, data.Engine, data.RedisClient)
	collectionService = service.NewCollectionService(collectionRepo, collectionGroupRepo, questionCommon)
	externalNotificationService = notification.NewExternalNotificationService(data.Engine, data.RedisClient, userNotificationConfigRepo, followRepo, emailService, userRepo, externalNotificationQueueService)
	answerActivityService = activity2.NewAnswerActivityService(answerActivityRepo)
	questionService = service.NewQuestionService(questionRepo, tagCommonService, questionCommon, userCommon, userRepo, revisionService, metaService, collectionCommon, answerActivityService, emailService, notificationQueueService, externalNotificationQueueService, activityQueueService, siteInfoCommonService, externalNotificationService)
	answerService = service.NewAnswerService(answerRepo, questionRepo, questionCommon, userCommon, collectionCommon, userRepo, revisionService, answerActivityService, answerCommon, voteRepo, emailService, userRoleRelService, notificationQueueService, externalNotificationQueueService, activityQueueService)
	searchParser = search_parser.NewSearchParser(tagCommonService, userCommon)
	searchService = service.NewSearchService(searchParser, searchRepo)
	serviceRevisionService = service.NewRevisionService(revisionRepo, userCommon, questionCommon, answerService, objService, questionRepo, answerRepo, tagRepo, tagCommonService, notificationQueueService, activityQueueService)
	reportHandle = report_handle_admin.NewReportHandle(questionCommon, commentRepo, notificationQueueService)
	reportAdminService = report_admin.NewReportAdminService(reportRepo, userCommon, answerRepo, questionRepo, commentCommonRepo, reportHandle, objService)
	userAdminService = user_admin.NewUserAdminService(userAdminRepo, userRoleRelService, authService, userCommon, userActiveActivityRepo, siteInfoCommonService, emailService, questionRepo, answerRepo, commentCommonRepo)
	reasonService = reason2.NewReasonService(reasonRepo)
	siteInfoService = siteinfo.NewSiteInfoService(siteInfoRepo, siteInfoCommonService, emailService, tagCommonService, questionCommon)
	notificationCommon = notificationcommon.NewNotificationCommon(data.Engine, data.RedisClient, notificationRepo, userCommon, activityRepo, followRepo, objService, notificationQueueService)
	notificationService = notification.NewNotificationService(data.Engine, data.RedisClient, notificationRepo, notificationCommon, revisionService, userRepo)
	activityCommon = activity_common2.NewActivityCommon(activityRepo, activityQueueService)
	commentCommonService = comment_common.NewCommentCommonService(commentCommonRepo)
	pluginCommonService = plugin_common.NewPluginCommonService(pluginConfigRepo, data.Engine, data.RedisClient)
	uploaderService = uploader.NewUploaderService(serviceConf, siteInfoCommonService)
	dashboardService = dashboard.NewDashboardService(questionRepo, answerRepo, commentCommonRepo, voteRepo, userRepo, reportRepo, siteInfoCommonService, serviceConf, data.Engine, data.RedisClient)
	activityService = activity2.NewActivityService(activityActivityRepo, userCommon, activityCommon, tagCommonService, objService, commentCommonService, revisionService, metaService)
}
