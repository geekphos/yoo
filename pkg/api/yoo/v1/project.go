package v1

import (
	"time"
)

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	BuildCmd    string `json:"build_cmd" binding:"required"`
	Dist        string `json:"dist" binding:"required"`
}

type CreateProjectResponse struct {
	ID          int32     `json:"id"`
	UserID      int32     `json:"user_id"`
	Pid         int32     `json:"pid"`
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

type GetProjectResponse struct {
	ID          int32     `json:"id"`
	UserID      int32     `json:"user_id"`
	Pid         int32     `json:"pid"`
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

type ListProjectRequest struct {
	Page        int    `json:"page" form:"page,default=1" binding:"required"`
	PageSize    int    `json:"page_size" form:"page_size,default=10" binding:"required"`
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
}

type ListProjectResponse struct {
	ID          int32     `json:"id"`
	UserID      int32     `json:"user_id"`
	Pid         int       `json:"pid"`
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
	ID          int32   `json:"id" uri:"id" binding:"required"`
	Name        *string `json:"name"`
	SSHURL      *string `json:"ssh_url"`
	HTTPURL     *string `json:"http_url"`
	WebURL      *string `json:"web_url"`
	BuildCmd    *string `json:"build_cmd"`
	Dist        *string `json:"dist"`
	Description *string `json:"description"`
}
