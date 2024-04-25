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
	"github.com/lawyer/repo/limit"
	"github.com/lawyer/repo/meta"
	notification2 "github.com/lawyer/repo/notification"
	"github.com/lawyer/repo/plugin_config"
	"github.com/lawyer/repo/question"
	"github.com/lawyer/repo/reason"
	"github.com/lawyer/repo/report"
	"github.com/lawyer/repo/revision"
	"github.com/lawyer/repo/role"
	"github.com/lawyer/repo/search_common"
	"github.com/lawyer/repo/site_info"
	"github.com/lawyer/repo/tag"
	"github.com/lawyer/repo/tag_common"
	"github.com/lawyer/repo/user"
	"github.com/lawyer/repo/user_external_login"
	"github.com/lawyer/repo/user_notification_config"
	"github.com/lawyer/repoCommon"
	"github.com/lawyer/service"
	"github.com/lawyer/service/action"
	sa "github.com/lawyer/service/activity"
	sact "github.com/lawyer/service/activity"
	ac "github.com/lawyer/service/activity_common"
	sac "github.com/lawyer/service/activity_common"
	sans "github.com/lawyer/service/answer_common"
	sau "github.com/lawyer/service/auth"
	scc "github.com/lawyer/service/collection_common"
	scom "github.com/lawyer/service/comment"
	scomc "github.com/lawyer/service/comment_common"
	se "github.com/lawyer/service/export"
	"github.com/lawyer/service/follow"
	smeta "github.com/lawyer/service/meta"
	sno "github.com/lawyer/service/notification_common"
	splug "github.com/lawyer/service/plugin_common"
	sq "github.com/lawyer/service/question_common"
	sr "github.com/lawyer/service/rank"
	srea "github.com/lawyer/service/reason_common"
	src "github.com/lawyer/service/report_common"
	srev "github.com/lawyer/service/revision"
	srole "github.com/lawyer/service/role"
	ssear "github.com/lawyer/service/search_common"
	"github.com/lawyer/service/siteinfo_common"
	stag "github.com/lawyer/service/tag_common"
	sur "github.com/lawyer/service/user_admin"
	usercommon "github.com/lawyer/service/user_common"
	suel "github.com/lawyer/service/user_external_login"
	sunc "github.com/lawyer/service/user_notification_config"
)

var (
	SiteInfoRepo               siteinfo_common.SiteInfoRepo
	AuthRepo                   sau.AuthRepo
	UserRepo                   usercommon.UserRepo
	ActivityRepo               ac.ActivityRepo
	UserRankRepo               sr.UserRankRepo
	UserActiveActivityRepo     sa.UserActiveActivityRepo
	EmailRepo                  se.EmailRepo
	UserRoleRelRepo            srole.UserRoleRelRepo
	RoleRepo                   srole.RoleRepo
	UserExternalLoginRepo      suel.UserExternalLoginRepo
	UserNotificationConfigRepo sunc.UserNotificationConfigRepo
	CaptchaRepo                action.CaptchaRepo
	CommentRepo                scom.CommentRepo
	AnswerRepo                 sans.AnswerRepo
	CommentCommonRepo          scomc.CommentCommonRepo
	QuestionRepo               sq.QuestionRepo
	TagCommonRepo              stag.TagCommonRepo
	TagRelRepo                 stag.TagRelRepo
	TagRepo                    stag.TagRepo
	RevisionRepo               srev.RevisionRepo
	VoteRepo                   sac.VoteRepo
	RolePowerRelRepo           srole.RolePowerRelRepo
	LimitRepo                  *limit.LimitRepo
	ReportRepo                 src.ReportRepo
	FollowRepo                 sac.FollowRepo
	FollowFollowRepo           follow.FollowRepo
	CollectionRepo             scc.CollectionRepo
	CollectionGroupRepo        service.CollectionGroupRepo
	MetaRepo                   smeta.MetaRepo
	AnswerActivityRepo         sact.AnswerActivityRepo
	ServiceVoteRepo            service.VoteRepo
	SearchRepo                 ssear.SearchRepo
	UserAdminRepo              sur.UserAdminRepo
	ReasonRepo                 srea.ReasonRepo
	NotificationRepo           sno.NotificationRepo
	ActivityActivityRepo       sact.ActivityRepo
	PluginConfigRepo           splug.PluginConfigRepo
)

func InitRepo() {

	SiteInfoRepo = site_info.NewSiteInfo()
	AuthRepo = auth.NewAuthRepo()
	UserRepo = user.NewUserRepo()
	ActivityRepo = repoCommon.NewActivityRepo()
	UserRankRepo = repoCommon.NewUserRankRepo()
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
	TagCommonRepo = tag_common.NewTagCommonRepo()
	TagRelRepo = tag.NewTagRelRepo()
	TagRepo = tag.NewTagRepo()
	RevisionRepo = revision.NewRevisionRepo()
	VoteRepo = activity_common.NewVoteRepo()
	RolePowerRelRepo = role.NewRolePowerRelRepo()
	LimitRepo = limit.NewRateLimitRepo()
	ReportRepo = report.NewReportRepo()
	FollowRepo = activity_common.NewFollowRepo()
	FollowFollowRepo = activity.NewFollowRepo()
	CollectionRepo = collection.NewCollectionRepo()
	CollectionGroupRepo = collection.NewCollectionGroupRepo()
	MetaRepo = meta.NewMetaRepo()
	AnswerActivityRepo = activity.NewAnswerActivityRepo()
	ServiceVoteRepo = activity.NewVoteRepo()
	SearchRepo = search_common.NewSearchRepo()
	UserAdminRepo = user.NewUserAdminRepo()
	ReasonRepo = reason.NewReasonRepo()
	NotificationRepo = notification2.NewNotificationRepo()
	ActivityActivityRepo = activity.NewActivityRepo()
	PluginConfigRepo = plugin_config.NewPluginConfigRepo()

}
