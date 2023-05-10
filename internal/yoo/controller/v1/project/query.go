package project

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/log"
	veldt "phos.cc/yoo/internal/pkg/validator"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

func (ctrl *ProjectController) Get(c *gin.Context) {

	log.C(c).Infow("GetByID project function called")

	// get id from the url
	var id int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}
	resp, err := ctrl.b.Projects().Get(c, int32(id))
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}

func (ctrl *ProjectController) List(c *gin.Context) {

	log.C(c).Infow("List project function called")

	var r v1.ListProjectRequest

	if err := c.ShouldBindQuery(&r); err != nil {
		if errs, _ := err.(validator.ValidationErrors); errs != nil {
			core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(veldt.Translate(errs)), nil)
			return
		} else {
			core.WriteResponse(c, errno.ErrBind, nil)
			return
		}
	}

	resp, total, err := ctrl.b.Projects().List(c, &r)
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

func (ctrl *ProjectController) All(c *gin.Context) {
	var r v1.ListProjectRequest

	if err := c.ShouldBindQuery(&r); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(veldt.Translate(errs)), nil)
		} else {
			core.WriteResponse(c, errno.ErrBind, nil)
		}
		return
	}

	resp, err := ctrl.b.Projects().All(c, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, gin.H{
		"data": resp,
		"code": 0,
	})
}
