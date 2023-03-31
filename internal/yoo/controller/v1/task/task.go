package task

import (
	"phos.cc/yoo/internal/yoo/biz"
	"phos.cc/yoo/internal/yoo/store"
)

type TaskController struct {
	b biz.Biz
}

func New(ds store.IStore) *TaskController {
	return &TaskController{b: biz.NewBiz(ds)}
}
