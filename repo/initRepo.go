package repo

import (
	"github.com/lawyer/repo/activity"
	"github.com/lawyer/repo/activity_common"
	"github.com/lawyer/repo/answer"
	"github.com/lawyer/repo/auth"
	"github.com/lawyer/repo/captcha"
	"github.com/lawyer/repo/collection"
	"github.com/lawyer/repo/comment"
	"github.com/lawyer/repo/export"
	"github.com/lawyer/repo/meta"
	"github.com/lawyer/repo/notification"
	"github.com/lawyer/repo/plugin_config"
	"github.com/lawyer/repo/question"
	"github.com/lawyer/repo/reason"
	"github.com/lawyer/repo/report"
	"github.com/lawyer/repo/revision"
	"github.com/lawyer/repo/role"
	"github.com/lawyer/repo/search_common"
	"github.com/lawyer/repo/tag"
	"github.com/lawyer/repo/user"
	"github.com/lawyer/repo/user_external_login"
	"github.com/lawyer/repo/user_notification_config"
)

var (
	AuthRepo *auth.AuthRepo
	UserRepo *user.UserRepo
	//ActivityRepo               ac.ActivityRepo
	//UserRankRepo               sr.UserRankRepo
	UserActiveActivityRepo     *activity.UserActiveActivityRepo
	EmailRepo                  *export.EmailRepo
	UserRoleRelRepo            *role.UserRoleRelRepo
	RoleRepo                   *role.RoleRepo
	UserExternalLoginRepo      *user_external_login.UserExternalLoginRepo
	UserNotificationConfigRepo *user_notification_config.UserNotificationConfigRepo
	CaptchaRepo                *captcha.CaptchaRepo
	CommentRepo                *comment.CommentRepo
	AnswerRepo                 *answer.AnswerRepo
	CommentCommonRepo          *comment.CommentRepo
	QuestionRepo               *question.QuestionRepo
	TagRepo                    *tag.TagRepo
	TagRelRepo                 *tag.TagRelRepo
	RevisionRepo               *revision.RevisionRepo
	RolePowerRelRepo           *role.RolePowerRelRepo
	//LimitRepo                  *limit.LimitRepo
	ReportRepo           *report.ReportRepo
	FollowRepo           *activity_common.FollowRepo
	FollowFollowRepo     *activity.FollowRepo
	CollectionRepo       *collection.CollectionRepo
	CollectionGroupRepo  *collection.CollectionGroupRepo
	MetaRepo             *meta.MetaRepo
	AnswerActivityRepo   *activity.AnswerActivityRepo
	VoteRepo             *activity.VoteRepo
	SearchRepo           *search_common.SearchRepo
	UserAdminRepo        *user.UserAdminRepo
	ReasonRepo           *reason.ReasonRepo
	NotificationRepo     *notification.NotificationRepo
	ActivityActivityRepo *activity.ActivityRepo
	PluginConfigRepo     *plugin_config.PluginConfigRepo
)

func InitRepo() {

	AuthRepo = auth.NewAuthRepo()
	UserRepo = user.NewUserRepo()
	//ActivityRepo = repoCommon.NewActivityRepo()
	//UserRankRepo = repoCommon.NewUserRankRepo()
	UserActiveActivityRepo = activity.NewUserActiveActivityRepo()
	EmailRepo = export.NewEmailRepo()

	UserRoleRelRepo = role.NewUserRoleRelRepo()

	RoleRepo = role.NewRoleRepo()
	UserExternalLoginRepo = user_external_login.NewUserExternalLoginRepo()
	UserNotificationConfigRepo = user_notification_config.NewUserNotificationConfigRepo()
	CaptchaRepo = captcha.NewCaptchaRepo()
	CommentRepo = comment.NewCommentRepo()
	CommentCommonRepo = comment.NewCommentCommonRepo()
	AnswerRepo = answer.NewAnswerRepo()
	QuestionRepo = question.NewQuestionRepo()
	TagRepo = tag.NewTagRepo()
	TagRelRepo = tag.NewTagRelRepo()
	RevisionRepo = revision.NewRevisionRepo()
	RolePowerRelRepo = role.NewRolePowerRelRepo()
	//LimitRepo = limit.NewRateLimitRepo()
	ReportRepo = report.NewReportRepo()
	FollowRepo = activity_common.NewFollowRepo()
	FollowFollowRepo = activity.NewFollowRepo()
	CollectionRepo = collection.NewCollectionRepo()
	CollectionGroupRepo = collection.NewCollectionGroupRepo()
	MetaRepo = meta.NewMetaRepo()
	AnswerActivityRepo = activity.NewAnswerActivityRepo()
	VoteRepo = activity.NewVoteRepo()
	SearchRepo = search_common.NewSearchRepo()
	UserAdminRepo = user.NewUserAdminRepo()
	ReasonRepo = reason.NewReasonRepo()
	NotificationRepo = notification.NewNotificationRepo()
	ActivityActivityRepo = activity.NewActivityRepo()
	PluginConfigRepo = plugin_config.NewPluginConfigRepo()

}
