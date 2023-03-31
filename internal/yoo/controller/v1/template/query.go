package template

import (
	"github.com/gin-gonic/gin"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/log"
	"strconv"
)

func (ctrl *TemplateController) Get(c *gin.Context) {
	log.C(c).Infow("GetByEmail template function called")

	// get id from the url
	var id int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}
	resp, err := ctrl.b.Templates().Get(c, int32(id))
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}
