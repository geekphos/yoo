package project

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/datatypes"
	"io"
	"net/http"
	"phos.cc/yoo/internal/pkg/known"
	"regexp"
	"strings"

	"github.com/jinzhu/copier"

	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/model"
	"phos.cc/yoo/internal/yoo/store"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

type ProjectBiz interface {
	Create(ctx context.Context, r *v1.CreateProjectRequest) error
	Get(ctx context.Context, id int32) (*v1.GetProjectResponse, error)
	Update(ctx context.Context, r *v1.UpdateProjectRequest) error
	List(ctx context.Context, r *v1.ListProjectRequest) ([]*v1.ListProjectResponse, int64, error)
	Categories(ctx context.Context) ([]string, error)
	Tags(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, id int32) error
}

type projectBiz struct {
	ds store.IStore
}

var _ ProjectBiz = (*projectBiz)(nil)

func New(ds store.IStore) ProjectBiz {
	return &projectBiz{ds: ds}
}

func (b *projectBiz) Create(ctx context.Context, r *v1.CreateProjectRequest) error {
	var projectM = &model.ProjectM{}
	_ = copier.Copy(projectM, r)
	userID := int32((ctx.(*gin.Context)).GetInt(known.XUserIDKey))
	projectM.UserID = userID
	projectM.Tags = datatypes.JSON(`["` + strings.Join(r.Tags, ",") + `"]`)

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

func (b *projectBiz) Update(ctx context.Context, r *v1.UpdateProjectRequest) error {

	projectM, err := b.ds.Projects().Get(ctx, r.ID)
	if err != nil {
		return errno.ErrProjectNotFound
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

	if r.Tag != "" {
		projectM.Tags = datatypes.JSON(`["` + r.Tag + `"]`)
	}

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

func (b *projectBiz) Categories(ctx context.Context) ([]string, error) {
	return b.ds.Projects().Categories(ctx)
}

func (b *projectBiz) Tags(ctx context.Context) ([]string, error) {
	return b.ds.Projects().Tags(ctx)
}

func (b *projectBiz) Delete(ctx context.Context, id int32) error {

	project, err := b.ds.Projects().Get(ctx, id)
	if err != nil {
		return errno.ErrProjectNotFound
	}

	// 删除gitlab上的项目
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v4/projects/%d", viper.GetString("gitlab-server"), project.Pid), nil)
	if err != nil {
		return err
	}

	req.Header.Add("PRIVATE-TOKEN", viper.GetString("gitlab-token"))

	resp, err := client.Do(req)
	if err != nil {
		return errno.InternalServerError
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return errno.InternalServerError
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errno.ErrProjectDelete.SetMessage(string(body))
	}

	if err := b.ds.Projects().Delete(ctx, id); err != nil {
		return errno.ErrProjectNotFound
	}

	return nil
}
