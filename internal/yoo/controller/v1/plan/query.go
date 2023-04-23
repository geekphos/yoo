package plan

import (
	"github.com/go-playground/validator/v10"
	veldt "phos.cc/yoo/internal/pkg/validator"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
	"strconv"

	"github.com/gin-gonic/gin"

	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
)

func (ctrl *PlanController) Get(c *gin.Context) {
	var id int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}

	resp, err := ctrl.b.Plans().Get(c, int32(id))

	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, resp)
}

func (ctrl *PlanController) List(c *gin.Context) {
	var r v1.ListPlanRequest

	if err := c.ShouldBindQuery(&r); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(veldt.Translate(errs)), nil)
		} else {
			core.WriteResponse(c, errno.ErrBind, nil)
		}
		return
	}

	resp, total, err := ctrl.b.Plans().List(c, &r)

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
