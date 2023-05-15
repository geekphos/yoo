package v1

import "time"

type CreateTaskRequest struct {
	PlanID    int32 `json:"plan_id" binding:"required,gte=1"`
	ProjectID int32 `json:"project_id" binding:"required,gte=1"`
}

type GetTaskResponse struct {
	ID           int32     `json:"id"`
	PlanID       int32     `json:"plan_id"`
	Status       int32     `json:"status"`
	Sha1         string    `json:"sha1"`          // 上次任务执行时，对应项目 git 的 sha1，用于比较是否需要重新打包
	FailedReason string    `json:"failed_reason"` // 打包失败的原因
	ProjectID    int32     `json:"project_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListTaskResponse struct {
	ID           int32     `json:"id"`
	PlanID       int32     `json:"plan_id"`
	Status       int32     `json:"status"`
	Sha1         string    `json:"sha1"`          // 上次任务执行时，对应项目 git 的 sha1，用于比较是否需要重新打包
	FailedReason string    `json:"failed_reason"` // 打包失败的原因
	ProjectID    int32     `json:"project_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdateTaskRequest struct {
	Status       int32   `json:"status"`
	Sha1         *string `json:"sha1"`          // 上次任务执行时，对应项目 git 的 sha1，用于比较是否需要重新打包
	FailedReason *string `json:"failed_reason"` // 打包失败的原因
}

type ListTaskRequest struct {
	Page      int32 `form:"page" binding:"required" default:"1"`
	PageSize  int   `form:"page_size" binding:"required" default:"10"`
	PlanID    int   `form:"plan_id" binding:"required"`
	Status    int   `form:"status" binding:"omitempty"`
	ProjectID int   `form:"project_id" binding:"gte=1,omitempty"`
}

type AllTaskRequest struct {
	PlanID int `form:"plan_id" binding:"required"`
	Status int `form:"status" binding:"omitempty"`
}

type AllTaskResponse struct {
	ID           int32     `json:"id"`
	PlanID       int32     `json:"plan_id"`
	Status       int32     `json:"status"`
	Sha1         string    `json:"sha1"`          // 上次任务执行时，对应项目 git 的 sha1，用于比较是否需要重新打包
	FailedReason string    `json:"failed_reason"` // 打包失败的原因
	ProjectID    int32     `json:"project_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
