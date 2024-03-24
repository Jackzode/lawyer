package permission

import (
	"context"
	"github.com/lawyer/commons/base/translator"
	entity2 "github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils"
)

// GetTagPermission get tag permission
func GetTagPermission(ctx context.Context, status int, canEdit, canDelete, canRecover bool) (
	actions []*schema.PermissionMemberAction) {
	lang := utils.GetLangByCtx(ctx)
	actions = make([]*schema.PermissionMemberAction, 0)
	if canEdit {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "edit",
			Name:   translator.Tr(lang, editActionName),
			Type:   "edit",
		})
	}

	if canDelete && status != entity2.TagStatusDeleted {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "delete",
			Name:   translator.Tr(lang, deleteActionName),
			Type:   "reason",
		})
	}

	if canRecover && status == entity2.QuestionStatusDeleted {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "undelete",
			Name:   translator.Tr(lang, undeleteActionName),
			Type:   "confirm",
		})
	}
	return actions
}

// GetTagSynonymPermission get tag synonym permission
func GetTagSynonymPermission(ctx context.Context, canEdit bool) (
	actions []*schema.PermissionMemberAction) {
	lang := utils.GetLangByCtx(ctx)
	actions = make([]*schema.PermissionMemberAction, 0)
	if canEdit {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "edit",
			Name:   translator.Tr(lang, editActionName),
			Type:   "edit",
		})
	}
	return actions
}
