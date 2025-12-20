package models

import "time"

type Professor struct {
	ID     uint   `gorm:"primaryKey"                    json:"id"`
	Name   string `gorm:"size:300;not null;uniqueIndex" json:"name"`
	UnitID uint   `gorm:"not null"                      json:"unit_id"`
	Usage  string `gorm:"size:200"                      json:"usage,omitempty"`
	Time   *int64 `                                     json:"time,omitempty"`

	Unit            Unit             `gorm:"foreignKey:UnitID"      json:"unit,omitempty"`
	ClassProfessors []ClassProfessor `gorm:"foreignKey:ProfessorID" json:"class_professors,omitempty"`
	CreatedAt       time.Time        `                              json:"created_at"`
	UpdatedAt       time.Time        `                              json:"updated_at"`
}