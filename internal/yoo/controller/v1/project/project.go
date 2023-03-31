package project

import (
	"phos.cc/yoo/internal/yoo/biz"
	"phos.cc/yoo/internal/yoo/store"
)

type ProjectController struct {
	b biz.Biz
}

func New(ds store.IStore) *ProjectController {
	return &ProjectController{b: biz.NewBiz(ds)}
}
