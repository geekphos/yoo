package user

import (
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/known"

	"github.com/gin-gonic/gin"

	"phos.cc/yoo/internal/pkg/log"
)

// Profile @Summary Profile
// @Description GetByEmail user profile
// @Tags User
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ProfileResponse
// @Security BearerAuth
// @Router /users/profile [get]
func (ctrl *UserController) Profile(c *gin.Context) {
	log.C(c).Infow("Profile function called")

	// get the user information by the user id from the context.
	email, _ := c.Get(known.XEmailKey)

	if resp, err := ctrl.b.Users().Profile(c, email.(string)); err != nil {
		core.WriteResponse(c, err, nil)
	} else {
		core.WriteResponse(c, nil, resp)
	}
}
