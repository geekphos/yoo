package template

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

func (ctrl *TemplateController) List(c *gin.Context) {
	log.C(c).Infow("List template function called")

	var r v1.ListTemplateRequest

	if err := c.ShouldBindQuery(&r); err != nil {
		if errs, _ := err.(validator.ValidationErrors); errs != nil {
			core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(veldt.Translate(errs)), nil)
			return
		}
	}

	resp, total, err := ctrl.b.Templates().List(c, &r)
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
