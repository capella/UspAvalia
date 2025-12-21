package models

type Unit struct {
	ID   uint   `gorm:"primaryKey"        json:"id"`
	Name string `gorm:"size:400;not null" json:"name"`

	Disciplines []Discipline `gorm:"foreignKey:UnitID" json:"disciplines,omitempty"`
	Professors  []Professor  `gorm:"foreignKey:UnitID" json:"professors,omitempty"`
}
