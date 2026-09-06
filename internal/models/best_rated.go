package models

// BestRated maps the Melhores view: class-professors with at least
// MinVotesForTopRated votes, in the order of database.TopRatedOrderSQL.
type BestRated struct {
	Average        float64 `gorm:"column:media"`
	VoteCount      int64   `gorm:"column:votos"`
	RecentVotes    int64   `gorm:"column:votos_recentes"`
	DisciplineName string  `gorm:"column:materia"`
	UnitName       string  `gorm:"column:unidade"`
	Code           string  `gorm:"column:codigo"`
	ProfessorName  string  `gorm:"column:professor"`
	ID             uint    `gorm:"column:id"`
}

func (BestRated) TableName() string {
	return "Melhores"
}
