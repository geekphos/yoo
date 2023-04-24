package template

import (
	"github.com/gin-gonic/gin"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/log"
	"strconv"
)

func (ctrl *TemplateController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete template function called")

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}

	if err := ctrl.b.Templates().Delete(c, int32(id)); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
