package models

import "time"

type ClassOffering struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	DisciplineID    uint   `gorm:"not null;index" json:"discipline_id"`
	Code            string `gorm:"size:20" json:"codigo"`
	TheoreticalCode string `gorm:"size:20" json:"codigo_teorica"`
	StartDate       string `gorm:"size:20" json:"inicio"` // Keep as string (DD/MM/YYYY)
	EndDate         string `gorm:"size:20" json:"fim"`
	Type            string `gorm:"size:50" json:"tipo"`
	Notes           string `gorm:"type:text" json:"observacoes"`
	Schedules       string `gorm:"type:text" json:"horario"` // JSON-encoded []HorarioInfo
	Vacancies       string `gorm:"type:text" json:"vagas"`   // JSON-encoded map[string]VagasInfo

	Discipline Discipline `gorm:"foreignKey:DisciplineID" json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
