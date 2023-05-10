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
	List(ctx context.Context, page, pageSize int, project *model.ProjectM) ([]*model.ProjectM, int64, error)
	All(ctx context.Context, project *model.ProjectM) ([]*model.ProjectM, error)
	Delete(ctx context.Context, id int32) error
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
	return p.db.WithContext(ctx).Save(project).Error
}

func (p *projects) List(ctx context.Context, page, pageSize int, project *model.ProjectM) ([]*model.ProjectM, int64, error) {
	var (
		projectMs []*model.ProjectM
		count     int64
	)

	query := p.db.WithContext(ctx).Model(&model.ProjectM{})

	if project.Name != "" {
		query = query.Where("name LIKE ?", "%"+project.Name+"%")
	}

	if project.Description != "" {
		query = query.Where("description LIKE ?", "%"+project.Description+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Find(&projectMs).Error; err != nil {
		return nil, 0, err
	}
	return projectMs, count, nil
}

func (p *projects) Delete(ctx context.Context, id int32) error {
	return p.db.WithContext(ctx).Delete(&model.ProjectM{}, id).Error
}

func (p *projects) All(ctx context.Context, project *model.ProjectM) ([]*model.ProjectM, error) {
	var projectMs []*model.ProjectM

	query := p.db.WithContext(ctx).Model(&model.ProjectM{})

	if project.Name != "" {
		query = query.Where("name LIKE ?", "%"+project.Name+"%")
	}

	if err := query.Find(&projectMs).Error; err != nil {
		return nil, err
	}
	return projectMs, nil
}
