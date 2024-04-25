package activity_common

import (
	"context"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/repoCommon"
	"time"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/uid"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

type ActivityRepo interface {
	GetActivityTypeByObjID(ctx context.Context, objectId string, action string) (activityType, rank int, hasRank int, err error)
	GetActivityTypeByObjectType(ctx context.Context, objectKey, action string) (activityType int, err error)
	GetActivity(ctx context.Context, session *xorm.Session, objectID, userID string, activityType int) (
		existsActivity *entity.Activity, exist bool, err error)
	GetUserIDObjectIDActivitySum(ctx context.Context, userID, objectID string) (int, error)
	GetActivityTypeByConfigKey(ctx context.Context, configKey string) (activityType int, err error)
	AddActivity(ctx context.Context, activity *entity.Activity) (err error)
	GetUsersWhoHasGainedTheMostReputation(
		ctx context.Context, startTime, endTime time.Time, limit int) (rankStat []*entity.ActivityUserRankStat, err error)
	GetUsersWhoHasVoteMost(
		ctx context.Context, startTime, endTime time.Time, limit int) (voteStat []*entity.ActivityUserVoteStat, err error)
}

type ActivityCommon struct {
}

// NewActivityCommon new activity common
func NewActivityCommon() *ActivityCommon {
	activity := &ActivityCommon{}
	services.ActivityQueueService.RegisterHandler(activity.HandleActivity)
	return activity
}

// HandleActivity handle activity message
func (ac *ActivityCommon) HandleActivity(ctx context.Context, msg *schema.ActivityMsg) error {
	activityType, err := repoCommon.NewActivityRepo().GetActivityTypeByConfigKey(ctx, string(msg.ActivityTypeKey))
	if err != nil {
		log.Errorf("error getting activity type %s, activity type is %d", err, activityType)
		return err
	}

	act := &entity.Activity{
		UserID:           msg.UserID,
		TriggerUserID:    msg.TriggerUserID,
		ObjectID:         uid.DeShortID(msg.ObjectID),
		OriginalObjectID: uid.DeShortID(msg.OriginalObjectID),
		ActivityType:     activityType,
		Cancelled:        entity.ActivityAvailable,
	}
	if len(msg.RevisionID) > 0 {
		act.RevisionID = converter.StringToInt64(msg.RevisionID)
	}
	if err = repoCommon.NewActivityRepo().AddActivity(ctx, act); err != nil {
		return err
	}
	return nil
}
