package project

import (
	"context"
	"regexp"

	"github.com/spf13/viper"
	"phos.cc/yoo/internal/yoo/gitlab"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"phos.cc/yoo/internal/pkg/known"

	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/model"
	"phos.cc/yoo/internal/yoo/store"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

type ProjectBiz interface {
	Create(ctx context.Context, r *v1.CreateProjectRequest) (*v1.CreateProjectResponse, error)
	Get(ctx context.Context, id int32) (*v1.GetProjectResponse, error)
	Update(ctx context.Context, r *v1.UpdateProjectRequest) error
	List(ctx context.Context, r *v1.ListProjectRequest) ([]*v1.ListProjectResponse, int64, error)
	All(ctx context.Context, r *v1.ListProjectRequest) ([]*v1.ListProjectResponse, error)
	Delete(ctx context.Context, id int32) error
}

type projectBiz struct {
	ds  store.IStore
	git gitlab.IGitlab
}

var _ ProjectBiz = (*projectBiz)(nil)

func New(ds store.IStore) ProjectBiz {
	return &projectBiz{ds: ds, git: gitlab.New(viper.GetString("gitlab.token"), viper.GetString("gitlab.server"), viper.GetInt32("gitlab.namespace"))}
}

func (b *projectBiz) Create(ctx context.Context, r *v1.CreateProjectRequest) (*v1.CreateProjectResponse, error) {

	// 在 gitlab 中创建项目
	repo, err := b.git.Create(ctx, &model.GitlabRepo{
		Name:        r.Name,
		Description: r.Description,
	})

	if err != nil {
		return nil, errno.ErrCreateRepoFail
	}

	var projectM = model.ProjectM{}
	_ = copier.Copy(&projectM, r)

	projectM.Pid = repo.ID
	projectM.SSHURL = repo.SSHURL
	projectM.HTTPURL = repo.HTTPURL
	projectM.WebURL = repo.WebURL

	userID := int32((ctx.(*gin.Context)).GetInt(known.XUserIDKey))
	projectM.UserID = userID

	if err := b.ds.Projects().Create(ctx, &projectM); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key '(name|repo|repo_id)'", err.Error()); match {
			return nil, errno.ErrProjectAlreadyExist
		}
		return nil, err
	}

	var resp v1.CreateProjectResponse

	_ = copier.Copy(&resp, &projectM)

	return &resp, nil
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

func (b *projectBiz) Update(ctx context.Context, r *v1.UpdateProjectRequest) error {

	projectM, err := b.ds.Projects().Get(ctx, r.ID)
	if err != nil {
		return errno.ErrProjectNotFound
	}

	var repo model.GitlabRepo
	repo.ID = projectM.Pid

	if r.Name != nil {
		repo.Name = *r.Name
	}
	if r.Description != nil {
		repo.Description = *r.Description
	}

	// 更新 git
	if err := b.git.Update(ctx, &repo); err != nil {
		return errno.ErrUpdateRepoFail
	}

	_ = copier.CopyWithOption(projectM, r, copier.Option{IgnoreEmpty: true})
	projectM.ID = r.ID

	if err := b.ds.Projects().Update(ctx, projectM); err != nil {
		return err
	}

	return nil
}

func (b *projectBiz) List(ctx context.Context, r *v1.ListProjectRequest) ([]*v1.ListProjectResponse, int64, error) {
	var projectM = &model.ProjectM{}
	_ = copier.Copy(projectM, r)

	projectMs, count, err := b.ds.Projects().List(ctx, r.Page, r.PageSize, projectM)
	if err != nil {
		return nil, 0, err
	}

	var userMap = make(map[int32]string)

	var resp []*v1.ListProjectResponse
	for _, projectM := range projectMs {
		if _, ok := userMap[projectM.UserID]; !ok {
			userM, err := b.ds.Users().Get(ctx, projectM.UserID)
			if err != nil {
				return nil, 0, errno.ErrUserNotFound
			}
			userMap[projectM.UserID] = userM.Nickname
		}

		var project = &v1.ListProjectResponse{}
		_ = copier.Copy(project, projectM)
		project.Username = userMap[projectM.UserID]

		resp = append(resp, project)
	}

	return resp, count, nil
}

func (b *projectBiz) Delete(ctx context.Context, id int32) error {

	p, err := b.ds.Projects().Get(ctx, id)
	if err != nil {
		return errno.ErrProjectNotFound
	}

	// 删除 gitlab 中的项目
	if err := b.git.Delete(ctx, p.Pid); err != nil {
		return errno.ErrRepoNotExist
	}

	if err := b.ds.Projects().Delete(ctx, id); err != nil {
		return errno.ErrProjectNotFound
	}

	return nil
}

func (b *projectBiz) All(ctx context.Context, r *v1.ListProjectRequest) ([]*v1.ListProjectResponse, error) {
	var projectM = &model.ProjectM{}
	_ = copier.Copy(projectM, r)

	projectMs, err := b.ds.Projects().All(ctx, projectM)
	if err != nil {
		return nil, errno.InternalServerError
	}

	var userMap = make(map[int32]string)

	var resp []*v1.ListProjectResponse
	for _, projectM := range projectMs {
		if _, ok := userMap[projectM.UserID]; !ok {
			userM, err := b.ds.Users().Get(ctx, projectM.UserID)
			if err != nil {
				return nil, errno.ErrUserNotFound
			}
			userMap[projectM.UserID] = userM.Nickname
		}

		var project = &v1.ListProjectResponse{}
		_ = copier.Copy(project, projectM)
		project.Username = userMap[projectM.UserID]

		resp = append(resp, project)
	}

	return resp, nil
}
