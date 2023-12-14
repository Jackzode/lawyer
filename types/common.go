package types

import "github.com/golang-jwt/jwt/v4"

type CustomClaim struct {
	jwt.RegisteredClaims
	UserName string
	Role     int
	Uid      string
}
