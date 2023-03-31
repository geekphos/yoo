package store

import (
	"context"

	"gorm.io/gorm"

	"phos.cc/yoo/internal/pkg/model"
)

type UserStore interface {
	Create(ctx context.Context, user *model.UserM) error
	GetByEmail(ctx context.Context, email string) (*model.UserM, error)
	Get(ctx context.Context, id int32) (*model.UserM, error)
	Update(ctx context.Context, user *model.UserM) error
}

type users struct {
	db *gorm.DB
}

var _ UserStore = (*users)(nil)

func newUsers(db *gorm.DB) *users {
	return &users{db: db}
}

func (u *users) Create(ctx context.Context, user *model.UserM) error {
	return u.db.WithContext(ctx).Create(user).Error
}

func (u *users) GetByEmail(ctx context.Context, email string) (*model.UserM, error) {
	var user = &model.UserM{}
	err := u.db.WithContext(ctx).Where("email = ?", email).First(user).Error
	return user, err
}

func (u *users) Update(ctx context.Context, user *model.UserM) error {
	return u.db.WithContext(ctx).Updates(user).Error
}

func (u *users) Get(ctx context.Context, id int32) (*model.UserM, error) {
	var user = &model.UserM{}
	err := u.db.WithContext(ctx).First(user, id).Error
	return user, err
}
