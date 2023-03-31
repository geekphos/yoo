package project

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/model"
	"phos.cc/yoo/internal/yoo/store"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

type ProjectBiz interface {
	Create(ctx context.Context, r *v1.CreateProjectRequest, id int32) error
	Get(ctx context.Context, id int32) (*v1.GetProjectResponse, error)
	Update(ctx context.Context, r *v1.UpdateProjectRequest, id int32) error
}

type projectBiz struct {
	ds store.IStore
}

var _ ProjectBiz = (*projectBiz)(nil)

func New(ds store.IStore) ProjectBiz {
	return &projectBiz{ds: ds}
}

func (b *projectBiz) Create(ctx context.Context, r *v1.CreateProjectRequest, id int32) error {
	var projectM = &model.ProjectM{}
	_ = copier.Copy(projectM, r)
	projectM.UserID = id

	if err := b.ds.Projects().Create(ctx, projectM); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key '(name|repo|repo_id)'", err.Error()); match {
			return errno.ErrProjectAlreadyExist
		}
		return err
	}

	return nil
}

func (b *projectBiz) Get(ctx context.Context, id int32) (*v1.GetProjectResponse, error) {
	projectM, err := b.ds.Projects().Get(ctx, id)
	if err != nil {
		return nil, errno.ErrProjectNotFound
	}

	userM, err := b.ds.Users().Get(ctx, projectM.UserID)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	var resp = &v1.GetProjectResponse{}
	_ = copier.Copy(resp, projectM)
	resp.Username = userM.Nickname

	return resp, nil
}

func (b *projectBiz) Update(ctx context.Context, r *v1.UpdateProjectRequest, id int32) error {

	projectM, err := b.ds.Projects().Get(ctx, id)
	if err != nil {
		return errno.ErrProjectNotFound
	}

	_ = copier.Copy(projectM, r)
	projectM.ID = id

	if err := b.ds.Projects().Update(ctx, projectM); err != nil {
		return err
	}

	return nil
}
