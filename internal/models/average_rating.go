package models

type AverageRating struct {
	ClassProfessorID uint    `gorm:"column:class_professor_id"`
	Average          float64 `gorm:"column:AVG(nota)"`
	VoteCount        int64   `gorm:"column:COUNT(*)"`
}

func (AverageRating) TableName() string {
	return "ListaMedias"
}