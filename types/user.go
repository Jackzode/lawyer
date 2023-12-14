package types

type UserInfo struct {
	ID             string `json:"id"`
	CreatedAt      int64  `json:"created_at"`
	LastLoginDate  int64  `json:"last_login_date"`
	Username       string `json:"username"`
	EMail          string `json:"e_mail"`
	AuthorityGroup int    `json:"authority_group"`
	DisplayName    string `json:"display_name"`
	//Avatar         string `json:"avatar"`
	Mobile   string `json:"mobile"`
	Bio      string `json:"bio"`
	Location string `json:"location"`
	//o admin  1 common 2 vip
	RoleID int `json:"role_id"`
	//0 ok , 1 封号 2 异常
	Status       int    `json:"status"`
	HavePassword bool   `json:"have_password"`
	PassWord     string `json:"passWord"`
}

type UserRegisterReq struct {
	Username string `json:"username" form:"username" validate:"required"`
	Email    string `json:"email" form:"email" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
	Captcha  string `json:"captcha" form:"captcha" validate:"required"`
	Uid      string `json:"uid" form:"uid" `
}

type UserEmailLoginReq struct {
	Email       string `validate:"required,e_mail" json:"e_mail" form:"e_mail"`
	Pass        string `validate:"required" json:"pass" form:"pass"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type LoginResponse struct {
	UserInfo UserInfo `json:"userInfo"`
	Token    string   `json:"token"`
}
