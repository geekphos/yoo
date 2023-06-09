package errno

var (
	// ErrProjectAlreadyExist is returned when the project already exists.
	ErrProjectAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.ProjectAlreadyExist", Message: "Project already exist."}

	// ErrProjectNotFound is returned when the project is not found.
	ErrProjectNotFound = &Errno{HTTP: 404, Code: "FailedOperation.ProjectNotFound", Message: "Project not found."}

	// ErrCompressBundles is returned when the compress bundles failed.
	ErrCompressBundles = &Errno{HTTP: 500, Code: "FailedOperation.CompressBundles", Message: "Compress bundles failed."}
	// ErrProjectDelete is returned when the project delete failed.
	ErrProjectDelete = &Errno{HTTP: 500, Code: "FailedOperation.ProjectDelete", Message: "Project delete failed."}

	// ErrProjectInUse is returned when the project is in use.
	ErrProjectInUse = &Errno{HTTP: 400, Code: "FailedOperation.ProjectInUse", Message: "Project in use."}
)
