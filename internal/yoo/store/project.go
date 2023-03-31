package store

import (
	"context"
	"gorm.io/gorm"
	"phos.cc/yoo/internal/pkg/model"
)

type ProjectStore interface {
	Create(ctx context.Context, project *model.ProjectM) error
	Get(ctx context.Context, id int32) (*model.ProjectM, error)
	Update(ctx context.Context, project *model.ProjectM) error
}

type projects struct {
	db *gorm.DB
}

var _ ProjectStore = (*projects)(nil)

func newProjects(db *gorm.DB) *projects {
	return &projects{db: db}
}

func (p *projects) Create(ctx context.Context, project *model.ProjectM) error {
	return p.db.WithContext(ctx).Create(project).Error
}

func (p *projects) Get(ctx context.Context, id int32) (*model.ProjectM, error) {
	var projectM = &model.ProjectM{}
	if err := p.db.WithContext(ctx).First(projectM, id).Error; err != nil {
		return nil, err
	}
	return projectM, nil
}

func (p *projects) Update(ctx context.Context, project *model.ProjectM) error {
	return p.db.WithContext(ctx).Updates(project).Error
}
