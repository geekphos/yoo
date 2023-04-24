package v1

import "time"

type CreateTemplateRequest struct {
	Name  string `json:"name" binding:"required"`
	Repo  string `json:"repo" binding:"required"`
	Brief string `json:"brief" binding:"required"`
}

type GetTemplateResponse struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Repo      string    `json:"repo"`
	Brief     string    `json:"brief"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListTemplateRequest struct {
	Page     int    `json:"page" form:"page,default=1" binding:"required"`
	PageSize int    `json:"page_size" form:"page_size,default=10" binding:"required"`
	Name     string `json:"name" form:"name"`
	Brief    string `json:"brief" form:"brief"`
}

type ListTemplateResponse struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Repo      string    `json:"repo"`
	Brief     string    `json:"brief"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AllTemplateRequest struct {
	Name  string `json:"name" form:"name"`
	Brief string `json:"brief" form:"brief"`
}

type AllTemplateResponse struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Repo      string    `json:"repo"`
	Brief     string    `json:"brief"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateTemplateRequest struct {
	ID    int     `json:"id" uri:"id" binding:"required"`
	Name  *string `json:"name"`
	Repo  *string `json:"repo"`
	Brief *string `json:"brief"`
}
