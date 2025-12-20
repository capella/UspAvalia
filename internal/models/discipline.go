package models

import "time"

type Discipline struct {
	ID     uint   `gorm:"primaryKey"        json:"id"`
	Name   string `gorm:"size:500;not null" json:"name"`
	Code   string `gorm:"size:40;not null"  json:"code"`
	UnitID uint   `gorm:"not null"          json:"unit_id"`
	Usage  string `gorm:"size:200"          json:"usage,omitempty"`
	Time   *int64 `                         json:"time,omitempty"`

	// MatrUSP additional fields
	Department   string `gorm:"size:200"  json:"departamento,omitempty"`
	Campus       string `gorm:"size:100"  json:"campus,omitempty"`
	CreditsClass int    `                 json:"creditos_aula,omitempty"`
	CreditsWork  int    `                 json:"creditos_trabalho,omitempty"`
	Objectives   string `gorm:"type:text" json:"objetivos,omitempty"`
	Summary      string `gorm:"type:text" json:"programa_resumido,omitempty"`

	Unit            Unit             `gorm:"foreignKey:UnitID"       json:"unit,omitempty"`
	ClassProfessors []ClassProfessor `gorm:"foreignKey:ClassID"      json:"class_professors,omitempty"`
	ClassOfferings  []ClassOffering  `gorm:"foreignKey:DisciplineID" json:"turmas,omitempty"`
	CreatedAt       time.Time        `                               json:"created_at"`
	UpdatedAt       time.Time        `                               json:"updated_at"`
}