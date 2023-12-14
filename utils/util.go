package utils

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid/v5"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/sony/sonyflake"
	"golang.org/x/crypto/bcrypt"
	"lawyer/common"
	"lawyer/types"
	"math/rand"
	"strconv"
	"time"
)

func ParseToken(tokenString string) (*types.CustomClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.CustomClaim{}, func(token *jwt.Token) (i interface{}, e error) {
		return secret, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*types.CustomClaim); ok && token.Valid {
			return claims, nil
		}
		return nil, TokenInvalid

	} else {
		return nil, TokenInvalid
	}
}

var secret = []byte("lawyer")
var (
	Issuer                 = "lawyer-test"
	ExpireDate             = time.Hour * 24 * 15
	ExpireBuffer     int64 = 1000 * 3600 * 24
	TokenExpired           = errors.New("token is expired")
	TokenNotValidYet       = errors.New("token not active yet")
	TokenMalformed         = errors.New("that's not even a token")
	TokenInvalid           = errors.New("token invalid")
)

func CreateToken(UserName, Uid string, Role int) (string, error) {

	claims := types.CustomClaim{
		UserName: UserName,
		Uid:      "12345",
		Role:     0,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"lawyer"},                     // 受众
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)),      // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ExpireDate)), // 过期时间 7天  配置文件
			Issuer:    Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func BcryptCheck(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func BcryptHash(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}

func CreateUUID() string {
	v4, err := uuid.NewV4()
	if err != nil {
		fmt.Println("createUUID...", err)
	}
	return v4.String()
}

func CreateCaptcha(length int) string {
	// 种子用于初始化随机数生成器
	//rand.Seed(time.Now().UnixNano())
	rand.NewSource(time.Now().UnixNano())
	charset := "0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = charset[rand.Intn(len(charset))]
	}
	//582846
	return string(code)
}

func GenerateTraceId() string {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		return common.DefaultTraceId
	}
	return strconv.FormatUint(id, 10)
}

func CreateUid() int64 {
	t := time.Now().Unix()
	captcha := CreateCaptcha(6)
	i, _ := strconv.ParseInt(captcha, 10, 32)
	return t + i
}
