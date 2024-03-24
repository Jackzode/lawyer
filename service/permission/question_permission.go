package permission

import (
	"context"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
)

// GetQuestionPermission get question permission
func GetQuestionPermission(ctx context.Context, userID string, creatorUserID string, status int,
	canEdit, canDelete, canClose, canReopen, canPin, canHide, canUnPin, canShow, canRecover bool) (
	actions []*schema.PermissionMemberAction) {
	lang := utils.GetLangByCtx(ctx)
	actions = make([]*schema.PermissionMemberAction, 0)
	if len(userID) > 0 {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "report",
			Name:   translator.Tr(lang, reportActionName),
			Type:   "reason",
		})
	}
	if (canEdit || userID == creatorUserID) && status != entity.QuestionStatusDeleted {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "edit",
			Name:   translator.Tr(lang, editActionName),
			Type:   "edit",
		})
	}
	if canClose && status == entity.QuestionStatusAvailable {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "close",
			Name:   translator.Tr(lang, closeActionName),
			Type:   "confirm",
		})
	}
	if canReopen {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "reopen",
			Name:   translator.Tr(lang, reopenActionName),
			Type:   "confirm",
		})
	}
	if canPin {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "pin",
			Name:   translator.Tr(lang, pinActionName),
			Type:   "confirm",
		})
	}
	if canHide {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "hide",
			Name:   translator.Tr(lang, hideActionName),
			Type:   "confirm",
		})
	}

	if canUnPin {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "unpin",
			Name:   translator.Tr(lang, unpinActionName),
			Type:   "confirm",
		})
	}

	if canShow {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "show",
			Name:   translator.Tr(lang, showActionName),
			Type:   "confirm",
		})
	}
	if canDelete || userID == creatorUserID {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "delete",
			Name:   translator.Tr(lang, deleteActionName),
			Type:   "confirm",
		})
	}

	if canRecover && status == entity.QuestionStatusDeleted {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "undelete",
			Name:   translator.Tr(lang, undeleteActionName),
			Type:   "confirm",
		})
	}
	return actions
}

// GetQuestionExtendsPermission get question extends permission
func GetQuestionExtendsPermission(ctx context.Context, canInviteOtherToAnswer bool) (
	actions []*schema.PermissionMemberAction) {
	lang := utils.GetLangByCtx(ctx)
	actions = make([]*schema.PermissionMemberAction, 0)
	if canInviteOtherToAnswer {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "invite_other_to_answer",
			Name:   translator.Tr(lang, inviteSomeoneToAnswerActionName),
			Type:   "confirm",
		})
	}
	return actions
}
