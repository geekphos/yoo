package v1

import "time"

type CreateTaskRequest struct {
	PlanID    int32 `json:"plan_id" binding:"required,gte=1"`
	ProjectID int32 `json:"project_id" binding:"required,gte=1"`
}

type GetTaskResponse struct {
	ID        int32     `json:"id"`
	PlanID    int32     `json:"plan_id"`
	Status    int32     `json:"status"`
	ProjectID int32     `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListTaskResponse struct {
	ID        int32     `json:"id"`
	PlanID    int32     `json:"plan_id"`
	Status    int32     `json:"status"`
	ProjectID int32     `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateTaskRequest struct {
	Status int32 `json:"status"`
}

type ListTaskRequest struct {
	Page      int `query:"page" binding:"required" default:"1"`
	PageSize  int `query:"page_size" binding:"required" default:"10"`
	PlanID    int `query:"plan_id" binding:"required"`
	Status    int `query:"status" binding:"omitempty"`
	ProjectID int `query:"project_id" binding:"gte=1,omitempty"`
}

type AllTaskRequest struct {
	PlanID int `query:"plan_id" binding:"required"`
	Status int `query:"status" binding:"omitempty"`
}

type AllTaskResponse struct {
	ID        int32     `json:"id"`
	PlanID    int32     `json:"plan_id"`
	Status    int32     `json:"status"`
	ProjectID int32     `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
