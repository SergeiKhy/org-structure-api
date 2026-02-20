package model

import (
	"time"
)

type Department struct {
	ID        int          `json:"id" gorm:"primaryKey"`
	Name      string       `json:"name" gorm:"size:200;not null"`
	ParentID  *int         `json:"parent_id" gorm:"index"`
	CreatedAt time.Time    `json:"created_at"`
	Employees []Employee   `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"`
	Children  []Department `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

type Employee struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	DepartmentID int        `json:"department_id" gorm:"not null;index"`
	FullName     string     `json:"full_name" gorm:"size:200;not null"`
	Position     string     `json:"position" gorm:"size:200;not null"`
	HiredAt      *time.Time `json:"hired_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// DTO для запросов
type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id"`
}

type CreateEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at"`
}

type UpdateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id"`
}
