package task

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
)

func (ctrl *TaskController) Delete(c *gin.Context) {

	ids := c.Param("pids")

	list := strings.Split(ids, ",")

	var idList []int32

	for _, v := range list {
		id, err := strconv.Atoi(v)
		if err != nil {
			core.WriteResponse(c, errno.ErrInvalidParameter, nil)
			return
		}
		idList = append(idList, int32(id))
	}

	if err := ctrl.b.Tasks().DeleteByPids(c, idList); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
