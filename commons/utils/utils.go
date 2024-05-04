package utils

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/utils/pager"
	"github.com/segmentfault/pacman/i18n"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

// GetEnableShortID get language from header
func GetEnableShortID(ctx context.Context) bool {
	flag, ok := ctx.Value(constant.ShortIDFlag).(bool)
	if ok {
		return flag
	}
	return false
}

// GetLang get language from header
func GetLang(ctx *gin.Context) i18n.Language {
	acceptLanguage := ctx.GetHeader(constant.AcceptLanguageFlag)
	if len(acceptLanguage) == 0 {
		return i18n.DefaultLanguage
	}
	return i18n.Language(acceptLanguage)
}

func GetTraceIdFromHeader(ctx *gin.Context) string {
	trace := ctx.GetHeader(constant.TraceID)
	return trace
}

// GetLangByCtx get language from header
func GetLangByCtx(ctx context.Context) i18n.Language {
	acceptLanguage, ok := ctx.Value(constant.AcceptLanguageFlag).(i18n.Language)
	if ok {
		return acceptLanguage
	}
	return i18n.DefaultLanguage
}

func GenerateTraceId() string {
	newUUID, _ := uuid.NewUUID()
	return newUUID.String()
}

type CommentQuery struct {
	pager.PageCond
	// object id
	ObjectID string
	// query condition
	QueryCond string
	// user id
	UserID string
}

func (c *CommentQuery) GetOrderBy() string {
	if c.QueryCond == "vote" {
		return "vote_count DESC,created_at ASC"
	}
	if c.QueryCond == "created_at" {
		return "created_at DESC"
	}
	return "created_at ASC"
}

func EncryptPassword(Pass string) (string, error) {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(Pass), bcrypt.DefaultCost)
	// This encrypted string can be saved to the database and can be used as password matching verification
	return string(hashPwd), err
}

// ExtractToken extract token from context
func ExtractToken(ctx *gin.Context) (token string) {
	token = ctx.GetHeader("Authorization")
	if len(token) == 0 {
		token = ctx.Query("Authorization")
	}
	return strings.TrimPrefix(token, "lawyer-")
}

func ParseToken(tokenString string) (*CustomClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaim{}, func(token *jwt.Token) (i interface{}, e error) {
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
	if token == nil || !token.Valid {
		return nil, TokenInvalid
	}
	if claims, ok := token.Claims.(*CustomClaim); ok {
		return claims, nil
	}
	return nil, TokenInvalid
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

type CustomClaim struct {
	jwt.RegisteredClaims
	UserName string
	Role     int
	Uid      string
}

func CreateToken(UserName, Uid string, Role int) (string, error) {

	claims := CustomClaim{
		UserName: UserName,
		Uid:      Uid,
		Role:     Role,
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
