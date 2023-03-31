package errno

var (
	// ErrTaskAlreadyExist is returned when the task already exists.
	ErrTaskAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.TaskAlreadyExist", Message: "Task already exists."}

	// ErrTaskNotFound is returned when the task is not found.
	ErrTaskNotFound = &Errno{HTTP: 404, Code: "FailedOperation.TaskNotFound", Message: "Task not found."}
)
