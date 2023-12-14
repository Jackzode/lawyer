package service

import "testing"

func TestUserService_SendEmailByVerification(t *testing.T) {

	u := &UserService{}
	u.SendEmailByVerification()

}
