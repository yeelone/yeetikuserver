package model

import (
	"time"
)

// Exam : Exam struct
type Exam struct {
	ID          uint64     `json:"id" gorm:"primary_key"`
	Creator     uint64     `json:"creator" gorm:"not null;"`
	Users       []User     `json:"users" gorm:"many2many:exam_users;"`
	Groups      []Group    `json:"groups" gorm:"many2many:exam_groups;"`
	Tags        []Tag      `json:"tags" gorm:"many2many:exam_tags;"`
	Questions   []Question `json:"questions" gorm:"many2many:exam_questions;"`
	Name        string     `json:"name" gorm:"not null;unique;"`
	Expired     bool       `json:"expired" gorm:"not null;"` //是否过期
	Description string     `json:"description"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
