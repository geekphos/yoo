package middleware

import (
	"github.com/gin-gonic/gin"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/known"
	"phos.cc/yoo/pkg/token"
	"time"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		email, id, tokenType, exp, err := token.ParseRequest(c)
		if err != nil || tokenType != token.AccessToken || exp.Before(time.Now()) {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)

			c.Abort()
			return
		}

		c.Set(known.XEmailKey, email)
		c.Set(known.XUserIDKey, id)
		c.Next()
	}
}
