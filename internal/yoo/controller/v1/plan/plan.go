package plan

import (
	"phos.cc/yoo/internal/yoo/biz"
	"phos.cc/yoo/internal/yoo/store"
)

type PlanController struct {
	b biz.Biz
}

func New(ds store.IStore) *PlanController {
	return &PlanController{b: biz.NewBiz(ds)}
}
