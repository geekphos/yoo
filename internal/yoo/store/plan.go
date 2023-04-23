package store

import (
	"context"
	"gorm.io/gorm"

	"phos.cc/yoo/internal/pkg/model"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

type PlanStore interface {
	Create(ctx context.Context, plan *model.PlanM) error
	Get(ctx context.Context, id int32) (*model.PlanM, error)
	List(ctx context.Context, r *v1.ListPlanRequest) ([]*model.PlanM, int64, error)
	Update(ctx context.Context, plan *model.PlanM) error
}

type plans struct {
	db *gorm.DB
}

var _ PlanStore = (*plans)(nil)

func newPlans(db *gorm.DB) PlanStore {
	return &plans{db: db}
}

func (p *plans) Create(ctx context.Context, plan *model.PlanM) error {
	return p.db.WithContext(ctx).Create(plan).Error
}

func (p *plans) Get(ctx context.Context, id int32) (*model.PlanM, error) {
	var planM = &model.PlanM{}
	if err := p.db.WithContext(ctx).First(planM, id).Error; err != nil {
		return nil, err
	}
	return planM, nil
}

func (p *plans) List(ctx context.Context, r *v1.ListPlanRequest) ([]*model.PlanM, int64, error) {
	var planMs []*model.PlanM
	var total int64
	query := p.db.WithContext(ctx).Model(&model.PlanM{})

	if r.Name != "" {
		query = query.Where("name LIKE ?", "%"+r.Name+"%")
	}

	if r.Status != 0 {
		query = query.Where("status = ?", r.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := p.db.WithContext(ctx).Find(&planMs).Error; err != nil {
		return nil, 0, err
	}
	return planMs, total, nil
}

func (p *plans) Update(ctx context.Context, plan *model.PlanM) error {
	return p.db.WithContext(ctx).Save(plan).Error
}
