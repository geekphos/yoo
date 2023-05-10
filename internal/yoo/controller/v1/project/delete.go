package project

import (
	"github.com/gin-gonic/gin"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"strconv"
)

func (ctrl *ProjectController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}

	// 查询是否有 task 在使用 project
	if task, err := ctrl.b.Tasks().GetByPid(c, int32(id)); err != nil {
		core.WriteResponse(c, err, nil)
		return
	} else if len(task) > 0 {
		core.WriteResponse(c, errno.ErrProjectInUse, nil)
		return
	}

	if err := ctrl.b.Projects().Delete(c, int32(id)); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
