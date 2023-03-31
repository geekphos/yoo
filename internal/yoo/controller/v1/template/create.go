package template

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/known"
	"phos.cc/yoo/internal/pkg/log"
	veldt "phos.cc/yoo/internal/pkg/validator"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

func (ctrl *TemplateController) Create(c *gin.Context) {
	log.C(c).Infow("Create template function called")

	var r v1.CreateTemplateRequest

	if err := c.ShouldBindJSON(&r); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(veldt.Translate(errs)), nil)
		} else {
			core.WriteResponse(c, errno.ErrBind, nil)
		}
		return
	}

	if err := ctrl.b.Templates().Create(c, &r, int32(c.GetInt(known.XUserIDKey))); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
