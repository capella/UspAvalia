package models

// BestRated maps the Melhores view: class-professors with at least
// MinEvaluatorsForTopRated evaluators, in the order of
// database.TopRatedOrderSQL. Movement is filled in by the handler from the
// top-rated snapshots and is not a view column.
type BestRated struct {
	Average          float64  `gorm:"column:media"`
	VoteCount        int64    `gorm:"column:votos"`
	Evaluators       int64    `gorm:"column:avaliadores"`
	RecentEvaluators int64    `gorm:"column:avaliadores_recentes"`
	DisciplineName   string   `gorm:"column:materia"`
	UnitName         string   `gorm:"column:unidade"`
	Code             string   `gorm:"column:codigo"`
	ProfessorName    string   `gorm:"column:professor"`
	ID               uint     `gorm:"column:id"`
	Movement         Movement `gorm:"-"`
}

func (BestRated) TableName() string {
	return "Melhores"
}
