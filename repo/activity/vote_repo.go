package activity

import (
	"context"
	"fmt"
	constant2 "github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity2 "github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/repo"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/log"
	"time"

	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/pkg/obj"
	"xorm.io/builder"

	"github.com/lawyer/commons/schema"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// VoteRepo activity repository
type VoteRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewVoteRepo new repository
func NewVoteRepo() *VoteRepo {
	return &VoteRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

func (vr *VoteRepo) Vote(ctx context.Context, op *schema.VoteOperationInfo) (err error) {
	noNeedToVote, err := vr.votePreCheck(ctx, op)
	if err != nil {
		return err
	}
	if noNeedToVote {
		return nil
	}

	sendInboxNotification := false
	maxDailyRank, err := repo.UserRankRepo.GetMaxDailyRank(ctx)
	if err != nil {
		return err
	}
	var userIDs []string
	for _, activity := range op.Activities {
		userIDs = append(userIDs, activity.ActivityUserID)
	}

	_, err = vr.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)

		userInfoMapping, err := vr.acquireUserInfo(session, userIDs)
		if err != nil {
			return nil, err
		}

		err = vr.setActivityRankToZeroIfUserReachLimit(ctx, session, op, userInfoMapping, maxDailyRank)
		if err != nil {
			return nil, err
		}

		sendInboxNotification, err = vr.saveActivitiesAvailable(session, op)
		if err != nil {
			return nil, err
		}

		err = vr.changeUserRank(ctx, session, op, userInfoMapping)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	for _, activity := range op.Activities {
		if activity.Rank == 0 {
			continue
		}
		vr.sendAchievementNotification(ctx, activity.ActivityUserID, op.ObjectCreatorUserID, op.ObjectID)
	}
	if sendInboxNotification {
		vr.sendVoteInboxNotification(ctx, op.OperatingUserID, op.ObjectCreatorUserID, op.ObjectID, op.VoteUp)
	}
	return nil
}

func (vr *VoteRepo) CancelVote(ctx context.Context, op *schema.VoteOperationInfo) (err error) {
	// Pre-Check
	// 1. check if the activity exist
	// 2. check if the activity is not cancelled
	// 3. if all activities are cancelled, return directly
	activities, err := vr.getExistActivity(ctx, op)
	if err != nil {
		return err
	}
	var userIDs []string
	for _, activity := range activities {
		if activity.Cancelled == entity2.ActivityCancelled {
			continue
		}
		userIDs = append(userIDs, activity.UserID)
	}
	if len(userIDs) == 0 {
		return nil
	}

	_, err = vr.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)

		userInfoMapping, err := vr.acquireUserInfo(session, userIDs)
		if err != nil {
			return nil, err
		}

		err = vr.cancelActivities(session, activities)
		if err != nil {
			return nil, err
		}

		err = vr.rollbackUserRank(ctx, session, activities, userInfoMapping)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	for _, activity := range activities {
		if activity.Rank == 0 {
			continue
		}
		vr.sendAchievementNotification(ctx, activity.UserID, op.ObjectCreatorUserID, op.ObjectID)
	}
	return nil
}

func (vr *VoteRepo) GetAndSaveVoteResult(ctx context.Context, objectID, objectType string) (
	up, down int64, err error) {
	up = vr.countVoteUp(ctx, objectID, objectType)
	down = vr.countVoteDown(ctx, objectID, objectType)
	err = vr.updateVotes(ctx, objectID, objectType, int(up-down))
	return
}

func (vr *VoteRepo) ListUserVotes(ctx context.Context, userID string,
	page int, pageSize int, activityTypes []int) (voteList []*entity2.Activity, total int64, err error) {
	session := vr.DB.Context(ctx)
	cond := builder.
		And(
			builder.Eq{"user_id": userID},
			builder.Eq{"cancelled": 0},
			builder.In("activity_type", activityTypes),
		)

	session.Where(cond).Desc("updated_at")

	total, err = pager.Help(page, pageSize, &voteList, &entity2.Activity{}, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (vr *VoteRepo) votePreCheck(ctx context.Context, op *schema.VoteOperationInfo) (noNeedToVote bool, err error) {
	activities, err := vr.getExistActivity(ctx, op)
	if err != nil {
		return false, err
	}
	done := 0
	for _, activity := range activities {
		if activity.Cancelled == entity2.ActivityAvailable {
			done++
		}
	}
	return done == len(op.Activities), nil
}

func (vr *VoteRepo) acquireUserInfo(session *xorm.Session, userIDs []string) (map[string]*entity2.User, error) {
	us := make([]*entity2.User, 0)
	err := session.In("id", userIDs).ForUpdate().Find(&us)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	users := make(map[string]*entity2.User, 0)
	for _, u := range us {
		users[u.ID] = u
	}
	return users, nil
}

func (vr *VoteRepo) setActivityRankToZeroIfUserReachLimit(ctx context.Context, session *xorm.Session,
	op *schema.VoteOperationInfo, userInfoMapping map[string]*entity2.User, maxDailyRank int) (err error) {
	// check if user reach daily rank limit
	for _, activity := range op.Activities {
		if activity.Rank > 0 {
			// check if reach max daily rank
			reach, err := repo.UserRankRepo.CheckReachLimit(ctx, session, activity.ActivityUserID, maxDailyRank)
			if err != nil {
				log.Error(err)
				return err
			}
			if reach {
				activity.Rank = 0
				continue
			}
		} else {
			// If user rank is lower than 1 after this action, then user rank will be set to 1 only.
			userCurrentScore := userInfoMapping[activity.ActivityUserID].Rank
			if userCurrentScore+activity.Rank < 1 {
				activity.Rank = 1 - userCurrentScore
			}
		}
	}
	return nil
}

func (vr *VoteRepo) changeUserRank(ctx context.Context, session *xorm.Session,
	op *schema.VoteOperationInfo,
	userInfoMapping map[string]*entity2.User) (err error) {
	for _, activity := range op.Activities {
		if activity.Rank == 0 {
			continue
		}
		user := userInfoMapping[activity.ActivityUserID]
		if user == nil {
			continue
		}
		if err = repo.UserRankRepo.ChangeUserRank(ctx, session,
			activity.ActivityUserID, user.Rank, activity.Rank); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func (vr *VoteRepo) rollbackUserRank(ctx context.Context, session *xorm.Session,
	activities []*entity2.Activity,
	userInfoMapping map[string]*entity2.User) (err error) {
	for _, activity := range activities {
		if activity.Rank == 0 {
			continue
		}
		user := userInfoMapping[activity.UserID]
		if user == nil {
			continue
		}
		if err = repo.UserRankRepo.ChangeUserRank(ctx, session,
			activity.UserID, user.Rank, -activity.Rank); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

// saveActivitiesAvailable save activities
// If activity not exist it will be created or else will be updated
// If this activity is already exist, set activity rank to 0
// So after this function, the activity rank will be correct for update user rank
func (vr *VoteRepo) saveActivitiesAvailable(session *xorm.Session, op *schema.VoteOperationInfo) (newAct bool, err error) {
	for _, activity := range op.Activities {
		existsActivity := &entity2.Activity{}
		exist, err := session.
			Where(builder.Eq{"object_id": op.ObjectID}).
			And(builder.Eq{"user_id": activity.ActivityUserID}).
			And(builder.Eq{"trigger_user_id": activity.TriggerUserID}).
			And(builder.Eq{"activity_type": activity.ActivityType}).
			Get(existsActivity)
		if err != nil {
			return false, err
		}
		if exist && existsActivity.Cancelled == entity2.ActivityAvailable {
			activity.Rank = 0
			continue
		}
		if exist {
			bean := &entity2.Activity{
				Cancelled: entity2.ActivityAvailable,
				Rank:      activity.Rank,
				HasRank:   activity.HasRank(),
			}
			session.Where("id = ?", existsActivity.ID)
			if _, err = session.Cols("`cancelled`", "`rank`", "`has_rank`").
				Update(bean); err != nil {
				return false, err
			}
		} else {
			insertActivity := entity2.Activity{
				ObjectID:         op.ObjectID,
				OriginalObjectID: op.ObjectID,
				UserID:           activity.ActivityUserID,
				TriggerUserID:    converter.StringToInt64(activity.TriggerUserID),
				ActivityType:     activity.ActivityType,
				Rank:             activity.Rank,
				HasRank:          activity.HasRank(),
				Cancelled:        entity2.ActivityAvailable,
			}
			_, err = session.Insert(&insertActivity)
			if err != nil {
				return false, err
			}
			newAct = true
		}
	}
	return newAct, nil
}

// cancelActivities cancel activities
// If this activity is already cancelled, set activity rank to 0
// So after this function, the activity rank will be correct for update user rank
func (vr *VoteRepo) cancelActivities(session *xorm.Session, activities []*entity2.Activity) (err error) {
	for _, activity := range activities {
		t := &entity2.Activity{}
		exist, err := session.ID(activity.ID).Get(t)
		if err != nil {
			log.Error(err)
			return err
		}
		if !exist {
			log.Error(fmt.Errorf("%s activity not exist", activity.ID))
			return fmt.Errorf("%s activity not exist", activity.ID)
		}
		//  If this activity is already cancelled, set activity rank to 0
		if t.Cancelled == entity2.ActivityCancelled {
			activity.Rank = 0
		}
		if _, err = session.ID(activity.ID).Cols("cancelled", "cancelled_at").
			Update(&entity2.Activity{
				Cancelled:   entity2.ActivityCancelled,
				CancelledAt: time.Now(),
			}); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func (vr *VoteRepo) getExistActivity(ctx context.Context, op *schema.VoteOperationInfo) ([]*entity2.Activity, error) {
	var activities []*entity2.Activity
	for _, action := range op.Activities {
		t := &entity2.Activity{}
		exist, err := vr.DB.Context(ctx).
			Where(builder.Eq{"user_id": action.ActivityUserID}).
			And(builder.Eq{"trigger_user_id": action.TriggerUserID}).
			And(builder.Eq{"activity_type": action.ActivityType}).
			And(builder.Eq{"object_id": op.ObjectID}).
			Get(t)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		if exist {
			activities = append(activities, t)
		}
	}
	return activities, nil
}

func (vr *VoteRepo) countVoteUp(ctx context.Context, objectID, objectType string) (count int64) {
	count, err := vr.countVote(ctx, objectID, objectType, constant2.ActVoteUp)
	if err != nil {
		log.Errorf("get vote up count error: %v", err)
	}
	return count
}

func (vr *VoteRepo) countVoteDown(ctx context.Context, objectID, objectType string) (count int64) {
	count, err := vr.countVote(ctx, objectID, objectType, constant2.ActVoteDown)
	if err != nil {
		log.Errorf("get vote down count error: %v", err)
	}
	return count
}

func (vr *VoteRepo) countVote(ctx context.Context, objectID, objectType, action string) (count int64, err error) {
	activity := &entity2.Activity{}
	activityType, _ := repo.ActivityRepo.GetActivityTypeByObjectType(ctx, objectType, action)
	count, err = vr.DB.Context(ctx).Where(builder.Eq{"object_id": objectID}).
		And(builder.Eq{"activity_type": activityType}).
		And(builder.Eq{"cancelled": 0}).
		Count(activity)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, err
}

func (vr *VoteRepo) updateVotes(ctx context.Context, objectID, objectType string, voteCount int) (err error) {
	session := vr.DB.Context(ctx)
	switch objectType {
	case constant2.QuestionObjectType:
		_, err = session.ID(objectID).Cols("vote_count").Update(&entity2.Question{VoteCount: voteCount})
	case constant2.AnswerObjectType:
		_, err = session.ID(objectID).Cols("vote_count").Update(&entity2.Answer{VoteCount: voteCount})
	case constant2.CommentObjectType:
		_, err = session.ID(objectID).Cols("vote_count").Update(&entity2.Comment{VoteCount: voteCount})
	}
	if err != nil {
		log.Error(err)
	}
	return
}

func (vr *VoteRepo) sendAchievementNotification(ctx context.Context, activityUserID, objectUserID, objectID string) {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return
	}

	_ = &schema.NotificationMsg{
		ReceiverUserID: activityUserID,
		TriggerUserID:  objectUserID,
		Type:           schema.NotificationTypeAchievement,
		ObjectID:       objectID,
		ObjectType:     objectType,
	}
	//todo
	//vr.notificationQueueService.Send(ctx, msg)
}

func (vr *VoteRepo) sendVoteInboxNotification(ctx context.Context, triggerUserID, receiverUserID, objectID string, upvote bool) {
	if triggerUserID == receiverUserID {
		return
	}
	objectType, _ := obj.GetObjectTypeStrByObjectID(objectID)

	msg := &schema.NotificationMsg{
		TriggerUserID:  triggerUserID,
		ReceiverUserID: receiverUserID,
		Type:           schema.NotificationTypeInbox,
		ObjectID:       objectID,
		ObjectType:     objectType,
	}
	if objectType == constant2.QuestionObjectType {
		if upvote {
			msg.NotificationAction = constant2.NotificationUpVotedTheQuestion
		} else {
			msg.NotificationAction = constant2.NotificationDownVotedTheQuestion
		}
	}
	if objectType == constant2.AnswerObjectType {
		if upvote {
			msg.NotificationAction = constant2.NotificationUpVotedTheAnswer
		} else {
			msg.NotificationAction = constant2.NotificationDownVotedTheAnswer
		}
	}
	if objectType == constant2.CommentObjectType {
		if upvote {
			msg.NotificationAction = constant2.NotificationUpVotedTheComment
		}
	}
	if len(msg.NotificationAction) > 0 {
		//vr.notificationQueueService.Send(ctx, msg)
	}
}
