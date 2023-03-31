package store

import (
	"context"
	"gorm.io/gorm"
	"phos.cc/yoo/internal/pkg/model"
)

type PlanStore interface {
	Create(ctx context.Context, plan *model.PlanM) error
	Get(ctx context.Context, id int32) (*model.PlanM, error)
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

func (p *plans) Update(ctx context.Context, plan *model.PlanM) error {
	return p.db.WithContext(ctx).Updates(plan).Error
}
