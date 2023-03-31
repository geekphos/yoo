package action

import (
	"phos.cc/yoo/internal/yoo/biz"
	"phos.cc/yoo/internal/yoo/store"
)

type ActionController struct {
	b biz.Biz
}

func New(ds store.IStore) *ActionController {
	return &ActionController{b: biz.NewBiz(ds)}
}
