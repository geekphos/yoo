package token

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type Config struct {
	key           string
	identityKey   string
	identityIDKey string
}

type TokenType string

const (
	AccessToken  TokenType = "access_token"
	RefreshToken TokenType = "refresh_token"
)

// ErrMissingHeader 表示 `Authorization` 请求头为空.
var ErrMissingHeader = errors.New("the length of the `Authorization` header is zero")

var (
	config = Config{"Rtg8BPKNEf2mB4mgvKONGPZZQSaJWNLijxR42qRgq0iBb5", "identityKey", "identityIDKey"}
	once   sync.Once
)

func Init(key string, identityKey string, identityIDKey string) {
	once.Do(func() {
		if key != "" {
			config.key = key
		}
		if identityKey != "" {
			config.identityKey = identityKey
		}
		if identityIDKey != "" {
			config.identityIDKey = identityIDKey
		}
	})
}

func Parse(tokenString string, key string) (string, int, TokenType, time.Time, error) {
	// 解析 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(key), nil
	})
	// 解析失败
	if err != nil {
		return "", 0, "", time.Time{}, err
	}

	var identityKey string
	var identityID int
	var tokenType TokenType
	var exp time.Time
	// 如果解析成功，获取 token 主题
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims[config.identityKey] == nil || claims[config.identityIDKey] == nil || claims["type"] == nil || claims["exp"] == nil {
			return "", 0, "", time.Time{}, errors.New("token claims error")
		}
		identityKey = claims[config.identityKey].(string)
		identityID = int(claims[config.identityIDKey].(float64))
		tokenType = TokenType(claims["type"].(string))
		exp = time.Unix(int64(claims["exp"].(float64)), 0)
	} else {
		return "", 0, "", time.Time{}, err
	}

	return identityKey, identityID, tokenType, exp, nil
}

func ParseRequest(c *gin.Context) (string, int, TokenType, time.Time, error) {
	header := c.Request.Header.Get("Authorization")

	if len(header) == 0 {
		return "", 0, "", time.Time{}, ErrMissingHeader
	}

	var t string

	fmt.Sscanf(header, "Bearer %s", &t)

	return Parse(t, config.key)
}

type Option func(claims *jwt.Token)

func WithExpDuration(duration time.Duration) Option {
	return func(claims *jwt.Token) {
		claims.Claims.(jwt.MapClaims)["exp"] = time.Now().Add(duration).Unix()
	}
}

func Sign(identity string, identityID int, tokenType TokenType, opts ...Option) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		config.identityKey:   identity,
		config.identityIDKey: identityID,
		"type":               tokenType,
		"nbf":                time.Now().Unix(),
		"iat":                time.Now().Unix(),
		"exp":                time.Now().Add(7 * 24 * time.Hour).Unix(),
	})

	for _, opt := range opts {
		opt(token)
	}

	return token.SignedString([]byte(config.key))
}
