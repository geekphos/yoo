package store

import (
	"context"

	"gorm.io/gorm"

	"phos.cc/yoo/internal/pkg/model"
)

type TaskStore interface {
	Create(ctx context.Context, task *model.TaskM) error
	Get(ctx context.Context, id int32) (*model.TaskM, error)
	Update(ctx context.Context, task *model.TaskM) error
	List(ctx context.Context, page, pageSize int, task *model.TaskM) ([]*model.TaskM, int64, error)
	All(ctx context.Context, task *model.TaskM) ([]*model.TaskM, error)
	Delete(ctx context.Context, id int32) error
}

type tasks struct {
	db *gorm.DB
}

var _ TaskStore = (*tasks)(nil)

func newTasks(db *gorm.DB) TaskStore {
	return &tasks{db: db}
}

func (p *tasks) Create(ctx context.Context, task *model.TaskM) error {
	return p.db.WithContext(ctx).Create(task).Error
}

func (p *tasks) Get(ctx context.Context, id int32) (*model.TaskM, error) {
	var taskM = &model.TaskM{}
	if err := p.db.WithContext(ctx).First(taskM, id).Error; err != nil {
		return nil, err
	}
	return taskM, nil
}

func (p *tasks) Update(ctx context.Context, task *model.TaskM) error {
	return p.db.WithContext(ctx).Updates(task).Error
}

func (p *tasks) List(ctx context.Context, page, pageSize int, task *model.TaskM) ([]*model.TaskM, int64, error) {
	var (
		taskMs []*model.TaskM
		count  int64
	)
	if err := p.db.WithContext(ctx).Model(&model.TaskM{}).Where(task).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := p.db.WithContext(ctx).Limit(pageSize).Offset((page - 1) * pageSize).Find(&taskMs).Error; err != nil {
		return nil, 0, err
	}
	return taskMs, count, nil
}

func (p *tasks) All(ctx context.Context, task *model.TaskM) ([]*model.TaskM, error) {
	var taskMs []*model.TaskM
	if err := p.db.WithContext(ctx).Where(task).Find(&taskMs).Error; err != nil {
		return nil, err
	}
	return taskMs, nil
}

func (p *tasks) Delete(ctx context.Context, id int32) error {
	return p.db.WithContext(ctx).Delete(&model.TaskM{}, id).Error
}
