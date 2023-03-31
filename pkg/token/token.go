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

func Parse(tokenString string, key string) (string, int, error) {
	// 解析 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(key), nil
	})
	// 解析失败
	if err != nil {
		return "", 0, err
	}

	var identityKey string
	var identityID int
	// 如果解析成功，获取 token 主题
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		identityKey = claims[config.identityKey].(string)
		identityID = int(claims[config.identityIDKey].(float64))
	} else {
		return "", 0, err
	}

	return identityKey, identityID, nil
}

func ParseRequest(c *gin.Context) (string, int, error) {
	header := c.Request.Header.Get("Authorization")

	if len(header) == 0 {
		return "", 0, ErrMissingHeader
	}

	var t string

	fmt.Sscanf(header, "Bearer %s", &t)

	return Parse(t, config.key)
}

func Sign(identity string, identityID int) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		config.identityKey:   identity,
		config.identityIDKey: identityID,
		"nbf":                time.Now().Unix(),
		"iat":                time.Now().Unix(),
		"exp":                time.Now().Add(100000 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(config.key))
}
