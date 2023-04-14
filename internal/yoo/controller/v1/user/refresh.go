package user

import (
	"time"

	"github.com/gin-gonic/gin"

	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/pkg/token"
)

func (ctrl *UserController) Refresh(c *gin.Context) {
	email, _, tokenType, exp, err := token.ParseRequest(c)
	if err != nil || tokenType != token.RefreshToken || exp.Before(time.Now()) {
		core.WriteResponse(c, errno.ErrTokenInvalid, nil)
		return
	}

	if resp, err := ctrl.b.Users().Refresh(c, email); err != nil {
		core.WriteResponse(c, err, nil)
	} else {
		core.WriteResponse(c, nil, resp)
	}
}
