package template

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/model"
	"phos.cc/yoo/internal/yoo/store"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

type TemplateBiz interface {
	Create(ctx context.Context, r *v1.CreateTemplateRequest, id int32) error
	Get(ctx context.Context, id int32) (*v1.GetTemplateResponse, error)
	List(ctx context.Context, r *v1.ListTemplateRequest) ([]*v1.ListTemplateResponse, int64, error)
	Update(ctx context.Context, r *v1.UpdateTemplateRequest) error
}

type templateBiz struct {
	ds store.IStore
}

var _ TemplateBiz = (*templateBiz)(nil)

func New(ds store.IStore) *templateBiz {
	return &templateBiz{ds: ds}
}

func (b *templateBiz) Create(ctx context.Context, r *v1.CreateTemplateRequest, id int32) error {
	var templateM = &model.TemplateM{}
	_ = copier.Copy(templateM, r)

	templateM.UserID = id

	if err := b.ds.Templates().Create(ctx, templateM); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key '(name|repo)'", err.Error()); match {
			return errno.ErrTemplateAlreadyExist
		}
		return err
	}

	return nil
}

func (b *templateBiz) Get(ctx context.Context, id int32) (*v1.GetTemplateResponse, error) {
	template, err := b.ds.Templates().Get(ctx, id)
	if err != nil {
		return nil, errno.ErrTemplateNotFound
	}

	user, err := b.ds.Users().Get(ctx, template.UserID)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	var resp = &v1.GetTemplateResponse{}
	_ = copier.Copy(resp, template)
	resp.Username = user.Nickname

	return resp, nil
}

func (b *templateBiz) List(ctx context.Context, r *v1.ListTemplateRequest) ([]*v1.ListTemplateResponse, int64, error) {
	var templateM = &model.TemplateM{}
	_ = copier.Copy(templateM, r)

	templates, count, err := b.ds.Templates().List(ctx, r.Page, r.PageSize, templateM)
	if err != nil {
		return nil, 0, err
	}

	var resp = make([]*v1.ListTemplateResponse, 0, len(templates))
	var userMap = make(map[int32]string)
	for _, template := range templates {
		var r = &v1.ListTemplateResponse{}
		_ = copier.Copy(r, template)
		if _, ok := userMap[template.UserID]; !ok {
			user, err := b.ds.Users().Get(ctx, template.UserID)
			if err != nil {
				return nil, 0, errno.InternalServerError
			}
			userMap[template.UserID] = user.Nickname
		}
		r.Username = userMap[template.UserID]
		resp = append(resp, r)
	}

	return resp, count, nil
}

func (b *templateBiz) Update(ctx context.Context, r *v1.UpdateTemplateRequest) error {
	templateM, err := b.ds.Templates().Get(ctx, int32(r.ID))
	if err != nil {
		return err
	}
	_ = copier.CopyWithOption(templateM, r, copier.Option{IgnoreEmpty: true})
	return b.ds.Templates().Update(ctx, templateM)
}
