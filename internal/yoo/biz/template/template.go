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
	template, user, err := b.ds.Templates().Get(ctx, id)
	if err != nil {
		return nil, errno.ErrTemplateNotFound
	}

	var resp = &v1.GetTemplateResponse{}
	_ = copier.Copy(resp, template)
	resp.Username = user.Nickname

	return resp, nil
}
