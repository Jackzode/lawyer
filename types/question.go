package types

import "time"

type QuestionReq struct {
	Title       string `form:"title" json:"title"`
	Content     string `form:"content" json:"content"`
	HTML        string `json:"-"`
	Tags        []*Tag `form:"tags" json:"tags"`
	UserID      string `json:"-"`
	CaptchaID   string `json:"captcha_id"` // captcha_id
	CaptchaCode string `json:"captcha_code"`
}

type Tag struct {
	// slug_name
	SlugName string `validate:"omitempty,gt=0,lte=35" json:"slug_name"`
	// display_name
	DisplayName string `validate:"omitempty,gt=0,lte=35" json:"display_name"`
	// original text
	OriginalText string `validate:"omitempty" json:"original_text"`
	// parsed text
	ParsedText string `json:"-"`
}

type QuestionPermission struct {
	// whether user can add it
	CanAdd                 bool `json:"-"`
	CanEdit                bool `json:"-"`
	CanDelete              bool `json:"-"`
	CanClose               bool `json:"-"`
	CanReopen              bool `json:"-"`
	CanPin                 bool `json:"-"`
	CanUnPin               bool `json:"-"`
	CanHide                bool `json:"-"`
	CanShow                bool `json:"-"`
	CanUseReservedTag      bool `json:"-"`
	CanInviteOtherToAnswer bool `json:"-"`
	CanAddTag              bool `json:"-"`
}

type Question struct {
	ID               string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt        time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt        time.Time `xorm:"updated_at TIMESTAMP"`
	UserID           string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	InviteUserID     string    `xorm:"TEXT invite_user_id"`
	LastEditUserID   string    `xorm:"not null default 0 BIGINT(20) last_edit_user_id"`
	Title            string    `xorm:"not null default '' VARCHAR(150) title"`
	OriginalText     string    `xorm:"not null MEDIUMTEXT original_text"`
	ParsedText       string    `xorm:"not null MEDIUMTEXT parsed_text"`
	Pin              int32     `xorm:"not null default 1 INT(11) pin"`
	Show             int32     `xorm:"not null default 1 INT(11) show"`
	Status           int32     `xorm:"not null default 1 INT(11) status"`
	ViewCount        int32     `xorm:"not null default 0 INT(11) view_count"`
	UniqueViewCount  int32     `xorm:"not null default 0 INT(11) unique_view_count"`
	VoteCount        int32     `xorm:"not null default 0 INT(11) vote_count"`
	AnswerCount      int32     `xorm:"not null default 0 INT(11) answer_count"`
	CollectionCount  int32     `xorm:"not null default 0 INT(11) collection_count"`
	FollowCount      int32     `xorm:"not null default 0 INT(11) follow_count"`
	AcceptedAnswerID string    `xorm:"not null default 0 BIGINT(20) accepted_answer_id"`
	LastAnswerID     string    `xorm:"not null default 0 BIGINT(20) last_answer_id"`
	PostUpdateTime   time.Time `xorm:"post_update_time TIMESTAMP"`
	RevisionID       string    `xorm:"not null default 0 BIGINT(20) revision_id"`
}
