package task

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/model"

	"phos.cc/yoo/internal/yoo/store"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

type TaskBiz interface {
	Create(ctx context.Context, r []*v1.CreateTaskRequest) error
	Update(ctx context.Context, r *v1.UpdateTaskRequest, id int32) error
	Get(ctx context.Context, id int32) (*v1.GetTaskResponse, error)
	GetByPid(ctx context.Context, pid int32) ([]*v1.GetTaskResponse, error)
	List(ctx context.Context, r *v1.ListTaskRequest) ([]*v1.ListTaskResponse, int64, error)
	All(ctx context.Context, r *v1.AllTaskRequest) ([]*v1.AllTaskResponse, error)
	DeleteByPids(ctx context.Context, ids []int32) error
}

type taskBiz struct {
	ds store.IStore
}

var _ TaskBiz = (*taskBiz)(nil)

func New(ds store.IStore) TaskBiz {
	return &taskBiz{ds: ds}
}

func (b *taskBiz) Create(ctx context.Context, r []*v1.CreateTaskRequest) error {
	var taskMs []*model.TaskM

	for _, v := range r {
		var taskM = &model.TaskM{}
		_ = copier.Copy(taskM, v)
		taskM.Status = 1
		taskMs = append(taskMs, taskM)
	}

	if err := b.ds.Tasks().Create(ctx, taskMs); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key '(plan_id|project_id)'", err.Error()); match {
			return errno.ErrTaskAlreadyExist
		}
		return err
	}

	return nil
}

func (b *taskBiz) Update(ctx context.Context, r *v1.UpdateTaskRequest, id int32) error {
	var taskM = &model.TaskM{}
	_ = copier.Copy(taskM, r)
	taskM.ID = id

	if err := b.ds.Tasks().Update(ctx, taskM); err != nil {
		return errno.InternalServerError
	}

	return nil
}

func (b *taskBiz) Get(ctx context.Context, id int32) (*v1.GetTaskResponse, error) {
	taskM, err := b.ds.Tasks().Get(ctx, id)
	if err != nil {
		return nil, errno.ErrTaskNotFound
	}

	var resp = &v1.GetTaskResponse{}
	_ = copier.Copy(resp, taskM)

	return resp, nil
}

func (b *taskBiz) GetByPid(ctx context.Context, pid int32) ([]*v1.GetTaskResponse, error) {
	taskMs, err := b.ds.Tasks().GetByPid(ctx, pid)
	if err != nil {
		return nil, errno.ErrTaskNotFound
	}

	var resp []*v1.GetTaskResponse
	for _, taskM := range taskMs {
		var task = &v1.GetTaskResponse{}
		_ = copier.Copy(task, taskM)
		resp = append(resp, task)
	}

	return resp, nil
}

func (b *taskBiz) List(ctx context.Context, r *v1.ListTaskRequest) ([]*v1.ListTaskResponse, int64, error) {
	var resp []*v1.ListTaskResponse

	taskM := &model.TaskM{}
	_ = copier.Copy(taskM, r)

	taskMs, total, err := b.ds.Tasks().List(ctx, r.Page, r.PageSize, taskM)
	if err != nil {
		return nil, 0, errno.InternalServerError
	}

	for _, taskM := range taskMs {
		var task = &v1.ListTaskResponse{}
		_ = copier.Copy(task, taskM)
		resp = append(resp, task)
	}

	return resp, total, nil
}

func (b *taskBiz) All(ctx context.Context, r *v1.AllTaskRequest) ([]*v1.AllTaskResponse, error) {
	var resp []*v1.AllTaskResponse

	taskM := &model.TaskM{}
	_ = copier.Copy(taskM, r)

	taskMs, err := b.ds.Tasks().All(ctx, taskM)
	if err != nil {
		return nil, errno.InternalServerError
	}

	for _, taskM := range taskMs {
		var task = &v1.AllTaskResponse{}
		_ = copier.Copy(task, taskM)
		resp = append(resp, task)
	}

	return resp, nil
}

func (b *taskBiz) DeleteByPids(ctx context.Context, ids []int32) error {
	if err := b.ds.Tasks().DeleteByPids(ctx, ids); err != nil {
		return errno.ErrTaskNotFound
	}

	return nil
}
