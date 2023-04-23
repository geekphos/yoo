package store

import (
	"context"
	"encoding/json"
	"github.com/samber/lo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"phos.cc/yoo/internal/pkg/model"
)

type ProjectStore interface {
	Create(ctx context.Context, project *model.ProjectM) error
	Get(ctx context.Context, id int32) (*model.ProjectM, error)
	Update(ctx context.Context, project *model.ProjectM) error
	List(ctx context.Context, page, pageSize int, project *model.ProjectM) ([]*model.ProjectM, int64, error)
	Categories(ctx context.Context) ([]string, error)
	Tags(ctx context.Context) ([]string, error)
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
		query = query.Where("description LIKE ?", "%"+project.Name+"%")
	}

	if project.Category != "" {
		query = query.Where("category = ?", project.Category)
	}

	if project.Tags != nil {
		var tags []string
		if err := json.Unmarshal(project.Tags, &tags); err != nil {
			return nil, 0, err
		}
		for _, tag := range tags {
			query = query.Where(datatypes.JSONArrayQuery("tags").Contains(tag))
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Find(&projectMs).Error; err != nil {
		return nil, 0, err
	}
	return projectMs, count, nil
}

func (p *projects) Categories(ctx context.Context) ([]string, error) {
	var res []string
	if result := p.db.WithContext(ctx).Table("projects").Select([]string{"category"}).Scan(&res); result.Error != nil {
		return nil, result.Error
	}
	return lo.Union(res), nil
}

func (p *projects) Tags(ctx context.Context) ([]string, error) {
	var tags []datatypes.JSON
	if result := p.db.WithContext(ctx).Table("projects").Select([]string{"tags"}).Scan(&tags); result.Error != nil {
		return nil, result.Error
	}
	var res [][]string
	lo.ForEach(tags, func(item datatypes.JSON, _ int) {
		var list []string
		if err := json.Unmarshal(item, &list); err == nil {
			res = append(res, list)
		}
	})

	return lo.Union(lo.Flatten(res)), nil
}
