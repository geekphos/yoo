package v1

import "time"

type CreatePlanRequest struct {
	Name string `json:"name"`
}

type GetPlanResponse struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListPlanRequest struct {
	Page     int    `json:"page" form:"page,default=1" binding:"required"`
	PageSize int    `json:"page_size" form:"page_size,default=10" binding:"required"`
	Name     string `json:"name" form:"name"`
	Status   int    `json:"status" form:"status"`
}

type ListPlanResponse struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdatePlanRequest struct {
	Name   *string `json:"name"`
	Status int     `json:"status"`
}
