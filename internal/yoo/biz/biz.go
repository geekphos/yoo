package biz

import (
	"phos.cc/yoo/internal/yoo/biz/plan"
	"phos.cc/yoo/internal/yoo/biz/project"
	"phos.cc/yoo/internal/yoo/biz/task"
	"phos.cc/yoo/internal/yoo/biz/template"
	"phos.cc/yoo/internal/yoo/biz/user"
	"phos.cc/yoo/internal/yoo/store"
)

type Biz interface {
	Users() user.UserBiz
	Templates() template.TemplateBiz
	Projects() project.ProjectBiz
	Plans() plan.PlanBiz
	Tasks() task.TaskBiz
}

type biz struct {
	ds store.IStore
}

var _ Biz = (*biz)(nil)

// NewBiz returns a new biz.
func NewBiz(ds store.IStore) *biz {
	return &biz{ds: ds}
}

func (b *biz) Users() user.UserBiz {
	return user.New(b.ds)
}

func (b *biz) Templates() template.TemplateBiz {
	return template.New(b.ds)
}

func (b *biz) Projects() project.ProjectBiz {
	return project.New(b.ds)
}

func (b *biz) Plans() plan.PlanBiz {
	return plan.New(b.ds)
}

func (b *biz) Tasks() task.TaskBiz {
	return task.New(b.ds)
}
