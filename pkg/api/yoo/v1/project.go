package v1

import (
	"time"

	"gorm.io/datatypes"
)

type CreateProjectRequest struct {
	Name        string   `json:"name" binding:"required"`
	Category    string   `json:"category" binding:"required"`
	Tags        []string `json:"tags" binding:"required,dive,required"`
	SSHURL      string   `json:"ssh_url" binding:"required"`
	HTTPURL     string   `json:"http_url" binding:"required"`
	WebURL      string   `json:"web_url" binding:"required"`
	BuildCmd    string   `json:"build_cmd" binding:"required"`
	Dist        string   `json:"dist" binding:"required"`
	Description string   `json:"description" binding:"required"`
}

type GetProjectResponse struct {
	ID          int32          `json:"id"`
	UserID      int32          `json:"user_id"`
	Username    string         `json:"username"`
	Name        string         `json:"name"`
	Badge       string         `json:"badge"`
	Category    string         `json:"category"`
	Tags        datatypes.JSON `json:"tags"`
	SSHURL      string         `json:"ssh_url"`
	HTTPURL     string         `json:"http_url"`
	WebURL      string         `json:"web_url"`
	BuildCmd    string         `json:"build_cmd"`
	Dist        string         `json:"dist"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ListProjectRequest struct {
	Page        int    `json:"page" form:"page,default=1" binding:"required"`
	PageSize    int    `json:"page_size" form:"page_size,default=10" binding:"required"`
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	Category    string `json:"category" form:"category"`
	Tag         string `json:"tag" form:"tag"`
}

type ListProjectResponse struct {
	ID          int32          `json:"id"`
	UserID      int32          `json:"user_id"`
	Username    string         `json:"username"`
	Name        string         `json:"name"`
	Badge       string         `json:"badge"`
	Category    string         `json:"category"`
	Tags        datatypes.JSON `json:"tags"`
	SSHURL      string         `json:"ssh_url"`
	HTTPURL     string         `json:"http_url"`
	WebURL      string         `json:"web_url"`
	BuildCmd    string         `json:"build_cmd"`
	Dist        string         `json:"dist"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type UpdateProjectRequest struct {
	ID          int32    `json:"id" uri:"id" binding:"required"`
	Name        *string  `json:"name"`
	SSHURL      *string  `json:"ssh_url"`
	HTTPURL     *string  `json:"http_url"`
	WebURL      *string  `json:"web_url"`
	BuildCmd    *string  `json:"build_cmd"`
	Dist        *string  `json:"dist"`
	Category    *string  `json:"category"`
	Tags        []string `json:"tags"`
	Description *string  `json:"description"`
}
