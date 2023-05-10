package model

type GitlabRepo struct {
	ID          int32  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	SSHURL      string `json:"ssh_url_to_repo,omitempty"`
	HTTPURL     string `json:"http_url_to_repo,omitempty"`
	WebURL      string `json:"web_url,omitempty"`
	Description string `json:"description,omitempty"`
}
