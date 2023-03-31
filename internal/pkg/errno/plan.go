package errno

var (
	// ErrPlanAlreadyExist plan already exist
	ErrPlanAlreadyExist = &Errno{HTTP: 400, Code: "FailedOperation.PlanAlreadyExist", Message: "Plan already exist."}

	// ErrPlanNotFound plan not found
	ErrPlanNotFound = &Errno{HTTP: 404, Code: "FailedOperation.PlanNotFound", Message: "Plan not found."}
)
