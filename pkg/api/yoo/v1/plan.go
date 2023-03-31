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
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdatePlanRequest struct {
	Name string `json:"name"`
}
