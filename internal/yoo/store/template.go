package store

import (
	"context"
	"gorm.io/gorm"
	"phos.cc/yoo/internal/pkg/model"
)

type TemplateStore interface {
	Create(ctx context.Context, template *model.TemplateM) error
	Get(ctx context.Context, id int32) (*model.TemplateM, error)
	List(ctx context.Context, page, pageSiz int, template *model.TemplateM) ([]*model.TemplateM, int64, error)
	Update(ctx context.Context, template *model.TemplateM) error
	Delete(ctx context.Context, id int32) error
}

type templates struct {
	db *gorm.DB
}

var _ TemplateStore = (*templates)(nil)

func newTemplates(db *gorm.DB) *templates {
	return &templates{db: db}
}

func (t *templates) Create(ctx context.Context, template *model.TemplateM) error {
	return t.db.WithContext(ctx).Create(template).Error
}

func (t *templates) Get(ctx context.Context, id int32) (*model.TemplateM, error) {
	var templateM = &model.TemplateM{}
	err := t.db.WithContext(ctx).First(templateM, id).Error
	return templateM, err
}

func (t *templates) List(ctx context.Context, page, pageSize int, template *model.TemplateM) ([]*model.TemplateM, int64, error) {
	var templates []*model.TemplateM
	var count int64

	query := t.db.WithContext(ctx).Model(&model.TemplateM{})
	if template.Name != "" {
		query = query.Where("name LIKE ?", "%"+template.Name+"%")
	}
	if template.Brief != "" {
		query = query.Where("brief LIKE ?", "%"+template.Brief+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, count, nil
}

func (t *templates) Update(ctx context.Context, template *model.TemplateM) error {
	return t.db.WithContext(ctx).Save(template).Error
}

func (t *templates) Delete(ctx context.Context, id int32) error {
	return t.db.WithContext(ctx).Delete(&model.TemplateM{}, id).Error
}
