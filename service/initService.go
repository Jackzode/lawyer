package service

import (
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/config"
	cc "github.com/lawyer/service/collection_common"
	"github.com/lawyer/service/comment_common"
	"github.com/lawyer/service/follow"
	"github.com/lawyer/service/meta"
	"github.com/lawyer/service/notice_queue"
	"github.com/lawyer/service/plugin_common"
	"github.com/lawyer/service/reason"
	"github.com/lawyer/service/revision_common"
	unc "github.com/lawyer/service/user_notification_config"
	"github.com/segmentfault/pacman/i18n"
)

var (
	AuthServicer                     *AuthService
	EmailServicer                    *EmailService
	RoleServicer                     *RoleService
	UserRoleRelServicer              *UserRoleRelService
	UserCommonServicer               *UserCommon
	UserNotificationConfigService    *unc.UserNotificationConfigService
	UserExternalLoginServicer        *UserExternalLoginService
	UserServicer                     *UserService
	CaptchaServicer                  *CaptchaService
	RevisionComServicer              *revision_common.RevisionService
	ActivityQueueServicer            ActivityQueueService
	ObjServicer                      *ObjService
	NotificationQueueService         notice_queue.NotificationQueueService
	ExternalNotificationQueueService notice_queue.ExternalNotificationQueueService
	CommentServicer                  *CommentService
	RolePowerRelServicer             *RolePowerRelService
	RankServicer                     *RankService
	ReportServicer                   *ReportService
	VoteServicer                     *VoteService
	TagServicer                      *TagService
	FollowService                    *follow.FollowService

	CollectionCommon             *cc.CollectionCommon
	AnswerCommonServicer         *AnswerCommon
	MetaService                  *meta.MetaService
	QuestionCommonServicer       *QuestionCommon
	CollectionServicer           *CollectionService
	ExternalNotificationServicer *ExternalNotificationService
	AnswerActivityServicer       *AnswerActivityService
	QuestionServicer             *QuestionService
	AnswerServicer               *AnswerService
	SearchParserServicer         *SearchParser
	SearchServicer               *SearchService
	RevisionServicer             *RevisionService
	ReportHandler                *ReportHandle
	ReportAdminServicer          *ReportAdminService
	UserAdminServicer            *UserAdminService
	ReasonService                *reason.ReasonService
	NotificationCommonServicer   *NotificationCommon
	NotificationServicer         *NotificationService
	ActivityCommonServicer       *ActivityCommon
	CommentCommonService         *comment_common.CommentCommonService
	PluginCommonService          *plugin_common.PluginCommonService
	UploaderServicer             UploaderService
	DashboardServicer            DashboardService
	ActivityServicer             *ActivityService
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

	AuthServicer = NewAuthService()
	EmailServicer = NewEmailService()
	RoleServicer = NewRoleService()
	UserRoleRelServicer = NewUserRoleRelService()
	UserCommonServicer = NewUserCommon()
	UserNotificationConfigService = unc.NewUserNotificationConfigService()
	UserExternalLoginServicer = NewUserExternalLoginService()
	UserServicer = NewUserService()
	CaptchaServicer = NewCaptchaService()
	RevisionComServicer = revision_common.NewRevisionService()
	ActivityQueueServicer = NewActivityQueueService()
	//TagCommonServicer = NewTagService()
	ObjServicer = NewObjService()

	NotificationQueueService = notice_queue.NewNotificationQueueService()
	ExternalNotificationQueueService = notice_queue.NewNewQuestionNotificationQueueService()
	CommentServicer = NewCommentService()

	RolePowerRelServicer = NewRolePowerRelService()
	RankServicer = NewRankService()

	ReportServicer = NewReportService()
	VoteServicer = NewVoteService()
	TagServicer = NewTagService()
	FollowService = follow.NewFollowService()

	CollectionCommon = cc.NewCollectionCommon()
	AnswerCommonServicer = NewAnswerCommon()
	MetaService = meta.NewMetaService()
	QuestionCommonServicer = NewQuestionCommon()
	CollectionServicer = NewCollectionService()
	ExternalNotificationServicer = NewExternalNotificationService()
	AnswerActivityServicer = NewAnswerActivityService()
	QuestionServicer = NewQuestionService()
	AnswerServicer = NewAnswerService()
	SearchParserServicer = NewSearchParser()
	SearchServicer = NewSearchService()
	RevisionServicer = NewRevisionService()
	ReportHandler = NewReportHandle()
	ReportAdminServicer = NewReportAdminService()
	UserAdminServicer = NewUserAdminService()
	ReasonService = reason.NewReasonService()
	//SiteInfoServicer = NewSiteInfoService()
	NotificationCommonServicer = NewNotificationCommon()
	NotificationServicer = NewNotificationService()
	ActivityCommonServicer = NewActivityCommon()
	CommentCommonService = comment_common.NewCommentCommonService()
	PluginCommonService = plugin_common.NewPluginCommonService()
	UploaderServicer = NewUploaderService()
	DashboardServicer = NewDashboardService()
	ActivityServicer = NewActivityService()
}
