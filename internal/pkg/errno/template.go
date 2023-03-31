package errno

var (
	// ErrTemplateAlreadyExist 代表模板已经存在.
	ErrTemplateAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.TemplateAlreadyExist", Message: "Template already exist."}

	// ErrTemplateNotFound 代表模板不存在.
	ErrTemplateNotFound = &Errno{HTTP: 404, Code: "FailedOperation.TemplateNotFound", Message: "Template not found."}
)
