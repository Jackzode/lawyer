/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package activity

import (
	"context"
	"fmt"
	constant2 "github.com/apache/incubator-answer/commons/constant"
	"github.com/apache/incubator-answer/commons/constant/reason"
	entity2 "github.com/apache/incubator-answer/commons/entity"
	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/log"
	"time"
	"xorm.io/builder"

	"github.com/apache/incubator-answer/commons/schema"
	"github.com/apache/incubator-answer/internal/service/activity"
	"github.com/apache/incubator-answer/internal/service/activity_common"
	"github.com/apache/incubator-answer/internal/service/notice_queue"
	"github.com/apache/incubator-answer/internal/service/rank"
	"github.com/apache/incubator-answer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// AnswerActivityRepo answer accepted
type AnswerActivityRepo struct {
	DB                       *xorm.Engine
	Cache                    *redis.Client
	activityRepo             activity_common.ActivityRepo
	userRankRepo             rank.UserRankRepo
	notificationQueueService notice_queue.NotificationQueueService
}

// NewAnswerActivityRepo new repository
func NewAnswerActivityRepo(
	DB *xorm.Engine,
	Cache *redis.Client,
	activityRepo activity_common.ActivityRepo,
	userRankRepo rank.UserRankRepo,
	notificationQueueService notice_queue.NotificationQueueService,
) activity.AnswerActivityRepo {
	return &AnswerActivityRepo{
		DB:                       DB,
		Cache:                    Cache,
		activityRepo:             activityRepo,
		userRankRepo:             userRankRepo,
		notificationQueueService: notificationQueueService,
	}
}

func (ar *AnswerActivityRepo) SaveAcceptAnswerActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) (
	err error) {
	// save activity
	_, err = ar.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)

		userInfoMapping, err := ar.acquireUserInfo(session, op.GetUserIDs())
		if err != nil {
			return nil, err
		}

		err = ar.saveActivitiesAvailable(session, op)
		if err != nil {
			return nil, err
		}

		err = ar.changeUserRank(ctx, session, op, userInfoMapping)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	// notification
	ar.sendAcceptAnswerNotification(ctx, op)
	return nil
}

func (ar *AnswerActivityRepo) SaveCancelAcceptAnswerActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) (
	err error) {
	// pre check
	activities, err := ar.getExistActivity(ctx, op)
	if err != nil {
		return err
	}
	var userIDs []string
	for _, act := range activities {
		if act.Cancelled == entity2.ActivityCancelled {
			continue
		}
		userIDs = append(userIDs, act.UserID)
	}
	if len(userIDs) == 0 {
		return nil
	}

	// save activity
	_, err = ar.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)

		userInfoMapping, err := ar.acquireUserInfo(session, userIDs)
		if err != nil {
			return nil, err
		}

		err = ar.cancelActivities(session, activities)
		if err != nil {
			return nil, err
		}

		err = ar.rollbackUserRank(ctx, session, activities, userInfoMapping)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	// notification
	ar.sendCancelAcceptAnswerNotification(ctx, op)
	return nil
}

func (ar *AnswerActivityRepo) acquireUserInfo(session *xorm.Session, userIDs []string) (map[string]*entity2.User, error) {
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

// saveActivitiesAvailable save activities
// If activity not exist it will be created or else will be updated
// If this activity is already exist, set activity rank to 0
// So after this function, the activity rank will be correct for update user rank
func (ar *AnswerActivityRepo) saveActivitiesAvailable(session *xorm.Session, op *schema.AcceptAnswerOperationInfo) (
	err error) {
	for _, act := range op.Activities {
		existsActivity := &entity2.Activity{}
		exist, err := session.
			Where(builder.Eq{"object_id": op.AnswerObjectID}).
			And(builder.Eq{"user_id": act.ActivityUserID}).
			And(builder.Eq{"trigger_user_id": act.TriggerUserID}).
			And(builder.Eq{"activity_type": act.ActivityType}).
			Get(existsActivity)
		if err != nil {
			return err
		}
		if exist && existsActivity.Cancelled == entity2.ActivityAvailable {
			act.Rank = 0
			continue
		}
		if exist {
			bean := &entity2.Activity{
				Cancelled: entity2.ActivityAvailable,
				Rank:      act.Rank,
				HasRank:   act.HasRank(),
			}
			session.Where("id = ?", existsActivity.ID)
			if _, err = session.Cols("`cancelled`", "`rank`", "`has_rank`").Update(bean); err != nil {
				return err
			}
		} else {
			insertActivity := entity2.Activity{
				ObjectID:         op.AnswerObjectID,
				OriginalObjectID: act.OriginalObjectID,
				UserID:           act.ActivityUserID,
				TriggerUserID:    converter.StringToInt64(act.TriggerUserID),
				ActivityType:     act.ActivityType,
				Rank:             act.Rank,
				HasRank:          act.HasRank(),
				Cancelled:        entity2.ActivityAvailable,
			}
			_, err = session.Insert(&insertActivity)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// cancelActivities cancel activities
// If this activity is already cancelled, set activity rank to 0
// So after this function, the activity rank will be correct for update user rank
func (ar *AnswerActivityRepo) cancelActivities(session *xorm.Session, activities []*entity2.Activity) (err error) {
	for _, act := range activities {
		t := &entity2.Activity{}
		exist, err := session.ID(act.ID).Get(t)
		if err != nil {
			log.Error(err)
			return err
		}
		if !exist {
			log.Error(fmt.Errorf("%s activity not exist", act.ID))
			return fmt.Errorf("%s activity not exist", act.ID)
		}
		//  If this activity is already cancelled, set activity rank to 0
		if t.Cancelled == entity2.ActivityCancelled {
			act.Rank = 0
		}
		if _, err = session.ID(act.ID).Cols("cancelled", "cancelled_at").
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

func (ar *AnswerActivityRepo) changeUserRank(ctx context.Context, session *xorm.Session,
	op *schema.AcceptAnswerOperationInfo,
	userInfoMapping map[string]*entity2.User) (err error) {
	for _, act := range op.Activities {
		if act.Rank == 0 {
			continue
		}
		user := userInfoMapping[act.ActivityUserID]
		if user == nil {
			continue
		}
		if err = ar.userRankRepo.ChangeUserRank(ctx, session,
			act.ActivityUserID, user.Rank, act.Rank); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func (ar *AnswerActivityRepo) rollbackUserRank(ctx context.Context, session *xorm.Session,
	activities []*entity2.Activity,
	userInfoMapping map[string]*entity2.User) (err error) {
	for _, act := range activities {
		if act.Rank == 0 {
			continue
		}
		user := userInfoMapping[act.UserID]
		if user == nil {
			continue
		}
		if err = ar.userRankRepo.ChangeUserRank(ctx, session,
			act.UserID, user.Rank, -act.Rank); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func (ar *AnswerActivityRepo) getExistActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) ([]*entity2.Activity, error) {
	var activities []*entity2.Activity
	for _, action := range op.Activities {
		var t []*entity2.Activity
		err := ar.DB.Context(ctx).
			Where(builder.Eq{"user_id": action.ActivityUserID}).
			And(builder.Eq{"activity_type": action.ActivityType}).
			And(builder.Eq{"object_id": op.AnswerObjectID}).
			Find(&t)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		if len(t) > 0 {
			activities = append(activities, t...)
		}
	}
	return activities, nil
}

func (ar *AnswerActivityRepo) sendAcceptAnswerNotification(
	ctx context.Context, op *schema.AcceptAnswerOperationInfo) {
	for _, act := range op.Activities {
		msg := &schema.NotificationMsg{
			Type:           schema.NotificationTypeAchievement,
			ObjectID:       op.AnswerObjectID,
			ReceiverUserID: act.ActivityUserID,
			TriggerUserID:  act.TriggerUserID,
		}
		msg.ObjectType = constant2.AnswerObjectType
		if msg.TriggerUserID != msg.ReceiverUserID {
			ar.notificationQueueService.Send(ctx, msg)
		}
	}

	for _, act := range op.Activities {
		msg := &schema.NotificationMsg{
			ReceiverUserID: act.ActivityUserID,
			Type:           schema.NotificationTypeInbox,
			ObjectID:       op.AnswerObjectID,
			TriggerUserID:  op.TriggerUserID,
		}
		if act.ActivityUserID != op.QuestionUserID {
			msg.ObjectType = constant2.AnswerObjectType
			msg.NotificationAction = constant2.NotificationAcceptAnswer
			ar.notificationQueueService.Send(ctx, msg)
		}
	}
}

func (ar *AnswerActivityRepo) sendCancelAcceptAnswerNotification(
	ctx context.Context, op *schema.AcceptAnswerOperationInfo) {
	for _, act := range op.Activities {
		msg := &schema.NotificationMsg{
			TriggerUserID:  act.TriggerUserID,
			ReceiverUserID: act.ActivityUserID,
			Type:           schema.NotificationTypeAchievement,
			ObjectID:       op.AnswerObjectID,
		}
		if act.ActivityUserID == op.QuestionObjectID {
			msg.ObjectType = constant2.QuestionObjectType
		} else {
			msg.ObjectType = constant2.AnswerObjectType
		}
		if msg.TriggerUserID != msg.ReceiverUserID {
			ar.notificationQueueService.Send(ctx, msg)
		}
	}
}
