// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameUserM = "users"

// UserM mapped from table <users>
type UserM struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Email     string    `gorm:"column:email;not null" json:"email"`
	Nickname  string    `gorm:"column:nickname;not null" json:"nickname"`
	Password  string    `gorm:"column:password;not null" json:"password"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Avatar    string    `gorm:"column:avatar;not null" json:"avatar"`
}

// TableName UserM's table name
func (*UserM) TableName() string {
	return TableNameUserM
}
