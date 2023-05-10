package errno

var (
	// ErrCreateRepoFail 创建项目失败
	ErrCreateRepoFail = &Errno{HTTP: 500, Code: "FailedOperation.CreateRepoFail", Message: "Create repo failed."}

	// ErrUpdateRepoFail 更新项目失败
	ErrUpdateRepoFail = &Errno{HTTP: 500, Code: "FailedOperation.UpdateRepoFail", Message: "Update repo failed."}

	// ErrRepoNotExist 项目不存在
	ErrRepoNotExist = &Errno{HTTP: 404, Code: "FailedOperation.RepoNotExist", Message: "Repo not exist."}
)
