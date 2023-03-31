package template

import (
	"phos.cc/yoo/internal/yoo/biz"
	"phos.cc/yoo/internal/yoo/store"
)

type TemplateController struct {
	b biz.Biz
}

func New(ds store.IStore) *TemplateController {
	return &TemplateController{b: biz.NewBiz(ds)}
}
