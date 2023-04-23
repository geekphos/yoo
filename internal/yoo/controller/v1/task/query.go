package task

import (
	"github.com/go-playground/validator/v10"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
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

func (ctrl *TaskController) List(c *gin.Context) {
	var r v1.ListTaskRequest

	if err := c.ShouldBindQuery(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	resp, total, err := ctrl.b.Tasks().List(c, &r)

	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, gin.H{
		"data": gin.H{
			"content": resp,
			"total":   total,
		},
		"code": 0,
	})
}

func (ctrl *TaskController) All(c *gin.Context) {

	var r v1.AllTaskRequest

	if err := c.ShouldBindQuery(&r); err != nil {
		if errs := err.(validator.ValidationErrors); errs != nil {
			core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(errs.Error()), nil)
		} else {
			core.WriteResponse(c, errno.ErrBind, nil)
		}
		return
	}

	resp, err := ctrl.b.Tasks().All(c, &r)

	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, gin.H{
		"data": resp,
		"code": 0,
	})
}
