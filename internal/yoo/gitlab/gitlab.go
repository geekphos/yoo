package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/model"
)

type IGitlab interface {
	Create(ctx context.Context, r *model.GitlabRepo) (*model.GitlabRepo, error)
	Update(ctx context.Context, r *model.GitlabRepo) error
	Delete(ctx context.Context, id int32) error
}

type gitlab struct {
	token      string
	server     string
	namespace  int32
	visibility string
}

var _ IGitlab = (*gitlab)(nil)

func New(token, server string, namespace int32) IGitlab {
	return &gitlab{token: token, server: server, namespace: namespace, visibility: "internal"}
}

func (g *gitlab) Create(ctx context.Context, r *model.GitlabRepo) (*model.GitlabRepo, error) {

	payload := map[string]interface{}{
		"name":         r.Name,
		"description":  r.Description,
		"namespace_id": g.namespace,
		"visibility":   g.visibility,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, errno.ErrInvalidParameter
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v4/projects", g.server), bytes.NewBuffer(body))
	if err != nil {
		return nil, errno.InternalServerError
	}
	req.Header.Set("PRIVATE-TOKEN", g.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errno.ErrCreateRepoFail
	}

	defer resp.Body.Close()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repo model.GitlabRepo

	if err := json.Unmarshal(all, &repo); err != nil {
		return nil, errno.ErrCreateRepoFail.SetMessage(err.Error())
	}

	return &repo, nil
}

func (g *gitlab) Update(ctx context.Context, r *model.GitlabRepo) error {
	client := &http.Client{}

	payload := map[string]interface{}{}
	if r.Name != "" {
		payload["name"] = r.Name
	}
	if r.Description != "" {
		payload["description"] = r.Description
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return errno.ErrInvalidParameter
	}

	// 如果没有需要更新的内容，直接返回
	if len(payload) == 0 {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/api/v4/projects/%d", g.server, r.ID), bytes.NewBuffer(body))
	if err != nil {
		return errno.InternalServerError
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errno.ErrUpdateRepoFail
	}
	defer resp.Body.Close()

	return nil
}

func (g *gitlab) Delete(ctx context.Context, id int32) error {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/api/v4/projects/%d", g.server, id), nil)
	if err != nil {
		return errno.ErrRepoNotExist
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errno.ErrProjectDelete
	}

	return nil
}
