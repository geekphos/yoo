// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNamePlanM = "plans"

// PlanM mapped from table <plans>
type PlanM struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UserID    int32     `gorm:"column:user_id;not null" json:"user_id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Status    int32     `gorm:"column:status;not null" json:"status"` // 计划状态: 1 空闲, 2 进行中, 3 失败, 4 成功
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// TableName PlanM's table name
func (*PlanM) TableName() string {
	return TableNamePlanM
}
