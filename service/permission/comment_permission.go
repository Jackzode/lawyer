package permission

import (
	"context"
	"github.com/lawyer/commons/base/handler"
	"github.com/lawyer/commons/base/translator"
	"github.com/lawyer/commons/constant"
	"time"

	"github.com/lawyer/commons/schema"
)

// GetCommentPermission get comment permission
func GetCommentPermission(ctx context.Context, userID string, creatorUserID string,
	createdAt time.Time, canEdit, canDelete bool) (actions []*schema.PermissionMemberAction) {
	lang := handler.GetLangByCtx(ctx)
	actions = make([]*schema.PermissionMemberAction, 0)
	if len(userID) > 0 {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "report",
			Name:   translator.Tr(lang, reportActionName),
			Type:   "reason",
		})
	}
	deadline := createdAt.Add(constant.CommentEditDeadline)
	if canEdit || (userID == creatorUserID && time.Now().Before(deadline)) {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "edit",
			Name:   translator.Tr(lang, editActionName),
			Type:   "edit",
		})
	}

	if canDelete || userID == creatorUserID {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "delete",
			Name:   translator.Tr(lang, deleteActionName),
			Type:   "reason",
		})
	}
	return actions
}
