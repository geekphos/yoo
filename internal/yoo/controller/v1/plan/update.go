package plan

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"strconv"

	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	veldt "phos.cc/yoo/internal/pkg/validator"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

func (ctrl *PlanController) Update(c *gin.Context) {
	var r *v1.UpdatePlanRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(veldt.Translate(errs)), nil)
			core.WriteResponse(c, errno.ErrBind, nil)
		}
		return
	}

	var id int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}

	if err := ctrl.b.Plans().Update(c, r, int32(id)); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
