package v1

type CreateResourceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Badge       *string `json:"badge"`
	Fake        bool    `json:"fake"`
	URL         *string `json:"url"`
}
