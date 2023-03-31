package task

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
)

func (ctrl *TaskController) Get(c *gin.Context) {
	var id int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}

	resp, err := ctrl.b.Tasks().Get(c, int32(id))

	if err != nil {
		core.WriteResponse(c, errno.ErrTaskNotFound, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}
