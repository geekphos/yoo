package project

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/log"
	veldt "phos.cc/yoo/internal/pkg/validator"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
	"strconv"
)

func (ctrl *ProjectController) Get(c *gin.Context) {

	log.C(c).Infow("GetByEmail project function called")

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

func (ctrl *ProjectController) Categories(c *gin.Context) {

	log.C(c).Infow("Categories project function called")

	resp, err := ctrl.b.Projects().Categories(c)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, gin.H{
		"data": resp,
		"code": 0,
	})
}

func (ctrl *ProjectController) Tags(c *gin.Context) {

	log.C(c).Infow("Tags project function called")

	resp, err := ctrl.b.Projects().Tags(c)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, gin.H{
		"data": resp,
		"code": 0,
	})
}
