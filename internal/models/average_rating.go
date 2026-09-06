package models

// MinEvaluatorsForTopRated is the minimum number of distinct evaluators
// (voters, excluding the difficulty vote type) a class-professor, professor
// or unit needs to appear in the top-rated lists (/destaques). Each
// evaluation writes four vote rows, so counting evaluators rather than rows
// stops one keen voter from qualifying an entry alone.
const MinEvaluatorsForTopRated = 4

// MinRecentEvaluatorsForTopRated is how many distinct evaluators from the
// last year an entry needs to compete for the top of the top-rated lists;
// entries below it are listed afterwards. See database.TopRatedOrderSQL.
const MinRecentEvaluatorsForTopRated = 3

// AverageRating maps the ListaMedias view. Average and VoteCount are the
// plain (legacy) aggregates; WeightedAverage and WeightSum use the recency
// weight from database.VoteWeightSQL, where a vote's weight halves for every
// full year since it was cast. Evaluators and RecentEvaluators count
// distinct voters overall and in the last year.
type AverageRating struct {
	ClassProfessorID uint    `gorm:"column:class_professor_id"`
	Average          float64 `gorm:"column:AVG(nota)"`
	VoteCount        int64   `gorm:"column:COUNT(*)"`
	WeightedAverage  float64 `gorm:"column:weighted_avg"`
	WeightSum        float64 `gorm:"column:weight_sum"`
	Evaluators       int64   `gorm:"column:evaluators"`
	RecentEvaluators int64   `gorm:"column:recent_evaluators"`
}

func (AverageRating) TableName() string {
	return "ListaMedias"
}
