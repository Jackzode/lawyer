package initServer

import (
	"github.com/apache/incubator-answer/initServer/data"
	"github.com/apache/incubator-answer/internal/repo/activity"
	"github.com/apache/incubator-answer/internal/repo/activity_common"
	"github.com/apache/incubator-answer/internal/repo/answer"
	"github.com/apache/incubator-answer/internal/repo/auth"
	"github.com/apache/incubator-answer/internal/repo/captcha"
	"github.com/apache/incubator-answer/internal/repo/collection"
	"github.com/apache/incubator-answer/internal/repo/comment"
	"github.com/apache/incubator-answer/internal/repo/export"
	"github.com/apache/incubator-answer/internal/repo/limit"
	"github.com/apache/incubator-answer/internal/repo/meta"
	notification2 "github.com/apache/incubator-answer/internal/repo/notification"
	"github.com/apache/incubator-answer/internal/repo/plugin_config"
	"github.com/apache/incubator-answer/internal/repo/question"
	"github.com/apache/incubator-answer/internal/repo/rank"
	"github.com/apache/incubator-answer/internal/repo/reason"
	"github.com/apache/incubator-answer/internal/repo/report"
	"github.com/apache/incubator-answer/internal/repo/revision"
	"github.com/apache/incubator-answer/internal/repo/role"
	"github.com/apache/incubator-answer/internal/repo/search_common"
	"github.com/apache/incubator-answer/internal/repo/site_info"
	"github.com/apache/incubator-answer/internal/repo/tag"
	"github.com/apache/incubator-answer/internal/repo/tag_common"
	"github.com/apache/incubator-answer/internal/repo/unique"
	"github.com/apache/incubator-answer/internal/repo/user"
	"github.com/apache/incubator-answer/internal/repo/user_external_login"
	"github.com/apache/incubator-answer/internal/repo/user_notification_config"
	"github.com/apache/incubator-answer/internal/service"
	"github.com/apache/incubator-answer/internal/service/action"
	sa "github.com/apache/incubator-answer/internal/service/activity"
	sact "github.com/apache/incubator-answer/internal/service/activity"
	ac "github.com/apache/incubator-answer/internal/service/activity_common"
	sac "github.com/apache/incubator-answer/internal/service/activity_common"
	sans "github.com/apache/incubator-answer/internal/service/answer_common"
	sau "github.com/apache/incubator-answer/internal/service/auth"
	scc "github.com/apache/incubator-answer/internal/service/collection_common"
	scom "github.com/apache/incubator-answer/internal/service/comment"
	scomc "github.com/apache/incubator-answer/internal/service/comment_common"
	se "github.com/apache/incubator-answer/internal/service/export"
	"github.com/apache/incubator-answer/internal/service/follow"
	smeta "github.com/apache/incubator-answer/internal/service/meta"
	sno "github.com/apache/incubator-answer/internal/service/notification_common"
	splug "github.com/apache/incubator-answer/internal/service/plugin_common"
	sq "github.com/apache/incubator-answer/internal/service/question_common"
	sr "github.com/apache/incubator-answer/internal/service/rank"
	srea "github.com/apache/incubator-answer/internal/service/reason_common"
	src "github.com/apache/incubator-answer/internal/service/report_common"
	srev "github.com/apache/incubator-answer/internal/service/revision"
	srole "github.com/apache/incubator-answer/internal/service/role"
	ssear "github.com/apache/incubator-answer/internal/service/search_common"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	stag "github.com/apache/incubator-answer/internal/service/tag_common"
	su "github.com/apache/incubator-answer/internal/service/unique"
	sur "github.com/apache/incubator-answer/internal/service/user_admin"
	usercommon "github.com/apache/incubator-answer/internal/service/user_common"
	suel "github.com/apache/incubator-answer/internal/service/user_external_login"
	sunc "github.com/apache/incubator-answer/internal/service/user_notification_config"
)

var (
	siteInfoRepo               siteinfo_common.SiteInfoRepo
	authRepo                   sau.AuthRepo
	userRepo                   usercommon.UserRepo
	uniqueIDRepo               su.UniqueIDRepo
	activityRepo               ac.ActivityRepo
	userRankRepo               sr.UserRankRepo
	userActiveActivityRepo     sa.UserActiveActivityRepo
	emailRepo                  se.EmailRepo
	userRoleRelRepo            srole.UserRoleRelRepo
	roleRepo                   srole.RoleRepo
	userExternalLoginRepo      suel.UserExternalLoginRepo
	userNotificationConfigRepo sunc.UserNotificationConfigRepo
	captchaRepo                action.CaptchaRepo
	commentRepo                scom.CommentRepo
	answerRepo                 sans.AnswerRepo
	commentCommonRepo          scomc.CommentCommonRepo
	questionRepo               sq.QuestionRepo
	tagCommonRepo              stag.TagCommonRepo
	tagRelRepo                 stag.TagRelRepo
	tagRepo                    stag.TagRepo
	revisionRepo               srev.RevisionRepo
	voteRepo                   sac.VoteRepo
	rolePowerRelRepo           srole.RolePowerRelRepo
	limitRepo                  *limit.LimitRepo
	reportRepo                 src.ReportRepo
	followRepo                 sac.FollowRepo
	followFollowRepo           follow.FollowRepo
	collectionRepo             scc.CollectionRepo
	collectionGroupRepo        service.CollectionGroupRepo
	metaRepo                   smeta.MetaRepo
	answerActivityRepo         sact.AnswerActivityRepo
	serviceVoteRepo            service.VoteRepo
	searchRepo                 ssear.SearchRepo
	userAdminRepo              sur.UserAdminRepo
	reasonRepo                 srea.ReasonRepo
	notificationRepo           sno.NotificationRepo
	activityActivityRepo       sact.ActivityRepo
	pluginConfigRepo           splug.PluginConfigRepo
)

func initRepo() {
	siteInfoRepo = site_info.NewSiteInfo(data.Engine, data.RedisClient)
	authRepo = auth.NewAuthRepo(data.Engine, data.RedisClient)
	userRepo = user.NewUserRepo(data.Engine, data.RedisClient)
	uniqueIDRepo = unique.NewUniqueIDRepo(data.Engine, data.RedisClient)
	activityRepo = activity_common.NewActivityRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	userRankRepo = rank.NewUserRankRepo(data.Engine, data.RedisClient)
	userActiveActivityRepo = activity.NewUserActiveActivityRepo(data.Engine, data.RedisClient, activityRepo, userRankRepo)
	emailRepo = export.NewEmailRepo(data.Engine, data.RedisClient)
	userRoleRelRepo = role.NewUserRoleRelRepo(data.Engine, data.RedisClient)
	roleRepo = role.NewRoleRepo(data.Engine, data.RedisClient)
	userExternalLoginRepo = user_external_login.NewUserExternalLoginRepo(data.Engine, data.RedisClient)
	userNotificationConfigRepo = user_notification_config.NewUserNotificationConfigRepo(data.Engine, data.RedisClient)
	captchaRepo = captcha.NewCaptchaRepo(data.Engine, data.RedisClient)
	commentRepo = comment.NewCommentRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	commentCommonRepo = comment.NewCommentCommonRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	answerRepo = answer.NewAnswerRepo(data.Engine, data.RedisClient, uniqueIDRepo, userRankRepo, activityRepo)
	questionRepo = question.NewQuestionRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	tagCommonRepo = tag_common.NewTagCommonRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	tagRelRepo = tag.NewTagRelRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	tagRepo = tag.NewTagRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	revisionRepo = revision.NewRevisionRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	voteRepo = activity_common.NewVoteRepo(data.Engine, data.RedisClient, activityRepo)
	rolePowerRelRepo = role.NewRolePowerRelRepo(data.Engine, data.RedisClient)
	limitRepo = limit.NewRateLimitRepo(data.Engine, data.RedisClient)
	reportRepo = report.NewReportRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	followRepo = activity_common.NewFollowRepo(data.Engine, data.RedisClient, uniqueIDRepo, activityRepo)
	followFollowRepo = activity.NewFollowRepo(data.Engine, data.RedisClient, uniqueIDRepo, activityRepo)
	collectionRepo = collection.NewCollectionRepo(data.Engine, data.RedisClient, uniqueIDRepo)
	collectionGroupRepo = collection.NewCollectionGroupRepo(data.Engine, data.RedisClient)
	metaRepo = meta.NewMetaRepo(data.Engine, data.RedisClient)
	answerActivityRepo = activity.NewAnswerActivityRepo(data.Engine, data.RedisClient, activityRepo, userRankRepo, notificationQueueService)
	serviceVoteRepo = activity.NewVoteRepo(data.Engine, data.RedisClient, activityRepo, userRankRepo, notificationQueueService)
	searchRepo = search_common.NewSearchRepo(data.Engine, data.RedisClient, uniqueIDRepo, userCommon, tagCommonService)
	userAdminRepo = user.NewUserAdminRepo(data.Engine, data.RedisClient, authRepo)
	reasonRepo = reason.NewReasonRepo()
	notificationRepo = notification2.NewNotificationRepo(data.Engine, data.RedisClient)
	activityActivityRepo = activity.NewActivityRepo(data.Engine, data.RedisClient)
	pluginConfigRepo = plugin_config.NewPluginConfigRepo(data.Engine, data.RedisClient)

}
