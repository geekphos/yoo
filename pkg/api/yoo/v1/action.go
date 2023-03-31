package v1

type ExecPlanRequest struct {
	PlanID  int32 `json:"plan_id" binding:"required"`
	GroupID int32 `json:"group_id" binding:"required"`
	IsMicro bool  `json:"is_micro"` // 是否使用微前端
}
