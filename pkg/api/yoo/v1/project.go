package v1

import "time"

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	SSHURL      string `json:"ssh_url" binding:"required"`
	HTTPURL     string `json:"http_url" binding:"required"`
	WebURL      string `json:"web_url" binding:"required"`
	BuildCmd    string `json:"build_cmd" binding:"required"`
	Dist        string `json:"dist" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type GetProjectResponse struct {
	ID          int32     `json:"id"`
	UserID      int32     `json:"user_id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	SSHURL      string    `json:"ssh_url"`
	HTTPURL     string    `json:"http_url"`
	WebURL      string    `json:"web_url"`
	BuildCmd    string    `json:"build_cmd"`
	Dist        string    `json:"dist"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	SSHURL      string `json:"ssh_url"`
	HTTPURL     string `json:"http_url"`
	WebURL      string `json:"web_url"`
	BuildCmd    string `json:"build_cmd"`
	Dist        string `json:"dist"`
	Description string `json:"description"`
}
