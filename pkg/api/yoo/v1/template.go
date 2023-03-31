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
