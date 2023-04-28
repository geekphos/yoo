package plan

import (
	"context"
	"github.com/jinzhu/copier"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/model"
	"phos.cc/yoo/internal/yoo/store"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
	"regexp"
)

type PlanBiz interface {
	Create(ctx context.Context, r *v1.CreatePlanRequest, id int32) error
	Get(ctx context.Context, id int32) (*v1.GetPlanResponse, error)
	List(ctx context.Context, r *v1.ListPlanRequest) ([]*v1.ListPlanResponse, int64, error)
	Update(ctx context.Context, r *v1.UpdatePlanRequest, id int32) error
	Delete(ctx context.Context, id int32) error
}

type planBiz struct {
	ds store.IStore
}

var _ PlanBiz = (*planBiz)(nil)

func New(ds store.IStore) PlanBiz {
	return &planBiz{ds: ds}
}

func (b *planBiz) Create(ctx context.Context, r *v1.CreatePlanRequest, id int32) error {
	var planM = &model.PlanM{}
	_ = copier.Copy(planM, r)
	planM.UserID = id
	planM.Status = 1

	if err := b.ds.Plans().Create(ctx, planM); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key 'name'", err.Error()); match {
			return errno.ErrPlanAlreadyExist
		}
		return err
	}

	return nil
}

func (b *planBiz) Get(ctx context.Context, id int32) (*v1.GetPlanResponse, error) {
	planM, err := b.ds.Plans().Get(ctx, id)
	if err != nil {
		return nil, errno.ErrPlanNotFound
	}

	userM, err := b.ds.Users().Get(ctx, planM.UserID)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	var resp = &v1.GetPlanResponse{}
	_ = copier.Copy(resp, planM)
	resp.Username = userM.Nickname

	return resp, nil
}

func (b *planBiz) List(ctx context.Context, r *v1.ListPlanRequest) ([]*v1.ListPlanResponse, int64, error) {
	planMs, total, err := b.ds.Plans().List(ctx, r)
	if err != nil {
		return nil, 0, err
	}

	var resp = make([]*v1.ListPlanResponse, 0, len(planMs))
	for _, planM := range planMs {
		userM, err := b.ds.Users().Get(ctx, planM.UserID)
		if err != nil {
			return nil, 0, errno.ErrUserNotFound
		}

		var respItem = &v1.ListPlanResponse{}
		_ = copier.Copy(respItem, planM)
		respItem.Username = userM.Nickname
		resp = append(resp, respItem)
	}

	return resp, total, nil
}

func (b *planBiz) Update(ctx context.Context, r *v1.UpdatePlanRequest, id int32) error {

	planM, err := b.ds.Plans().Get(ctx, id)
	if err != nil {
		return errno.ErrPlanNotFound
	}

	_ = copier.CopyWithOption(planM, r, copier.Option{IgnoreEmpty: true})
	planM.ID = id

	if err := b.ds.Plans().Update(ctx, planM); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key 'name'", err.Error()); match {
			return errno.ErrPlanAlreadyExist
		}
		return err
	}

	return nil
}

func (b *planBiz) Delete(ctx context.Context, id int32) error {
	if err := b.ds.Plans().Delete(ctx, id); err != nil {
		return errno.ErrPlanNotFound
	}

	return nil
}
