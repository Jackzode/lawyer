package schema

import (
	"encoding/json"
	"github.com/lawyer/commons/constant"
)

const (
	AccountActivationSourceType EmailSourceType = "account-activation"
	PasswordResetSourceType     EmailSourceType = "password-reset"
	ConfirmNewEmailSourceType   EmailSourceType = "password-reset"
	UnsubscribeSourceType       EmailSourceType = "unsubscribe"
	BindingSourceType           EmailSourceType = "binding"
)

type EmailSourceType string

type EmailCodeContent struct {
	SourceType EmailSourceType `json:"source_type"`
	Email      string          `json:"e_mail"`
	UserID     string          `json:"user_id"`
	// Used for unsubscribe notification
	NotificationSources []constant.NotificationSource `json:"notification_source,omitempty"`
	// Used for third-party login account binding
	BindingKey string `json:"binding_key,omitempty"`
}

func (r *EmailCodeContent) ToJSONString() string {
	codeBytes, _ := json.Marshal(r)
	return string(codeBytes)
}

func (r *EmailCodeContent) FromJSONString(data string) error {
	return json.Unmarshal([]byte(data), &r)
}

type RegisterTemplateData struct {
	SiteName    string
	RegisterUrl string
}

type PassResetTemplateData struct {
	SiteName     string
	PassResetUrl string
}

type ChangeEmailTemplateData struct {
	SiteName       string
	ChangeEmailUrl string
}

type TestTemplateData struct {
	SiteName string
}

type NewAnswerTemplateRawData struct {
	AnswerUserDisplayName string
	QuestionTitle         string
	QuestionID            string
	AnswerID              string
	AnswerSummary         string
	UnsubscribeCode       string
}

type NewAnswerTemplateData struct {
	SiteName       string
	DisplayName    string
	QuestionTitle  string
	AnswerUrl      string
	AnswerSummary  string
	UnsubscribeUrl string
}

type NewInviteAnswerTemplateRawData struct {
	InviterDisplayName string
	QuestionTitle      string
	QuestionID         string
	UnsubscribeCode    string
}

type NewInviteAnswerTemplateData struct {
	SiteName       string
	DisplayName    string
	QuestionTitle  string
	InviteUrl      string
	UnsubscribeUrl string
}

type NewCommentTemplateRawData struct {
	CommentUserDisplayName string
	QuestionTitle          string
	QuestionID             string
	AnswerID               string
	CommentID              string
	CommentSummary         string
	UnsubscribeCode        string
}

type NewCommentTemplateData struct {
	SiteName       string
	DisplayName    string
	QuestionTitle  string
	CommentUrl     string
	CommentSummary string
	UnsubscribeUrl string
}

type NewQuestionTemplateRawData struct {
	QuestionAuthorUserID string
	QuestionTitle        string
	QuestionID           string
	UnsubscribeCode      string
	Tags                 []string
	TagIDs               []string
}

type NewQuestionTemplateData struct {
	SiteName       string
	QuestionTitle  string
	QuestionUrl    string
	Tags           string
	UnsubscribeUrl string
}
