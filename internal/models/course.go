package models

import "time"

type Course struct {
	ID      uint   `gorm:"primaryKey"              json:"id"`
	Code    string `gorm:"size:20;not null;unique" json:"code"`
	Name    string `gorm:"size:500;not null"       json:"name"`
	UnitID  uint   `gorm:"not null"                json:"unit_id"`
	Period  string `gorm:"size:20"                 json:"period"`
	Periods string `gorm:"type:text"               json:"periods"` // JSON-encoded curriculum data

	Unit      Unit      `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	CreatedAt time.Time `                         json:"created_at"`
	UpdatedAt time.Time `                         json:"updated_at"`
}
