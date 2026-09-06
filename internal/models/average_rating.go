package models

// MinVotesForTopRated is the minimum number of votes (excluding the
// difficulty vote type) a class-professor, professor or unit needs to appear
// in the top-rated lists (/destaques).
const MinVotesForTopRated = 15

// AverageRating maps the ListaMedias view. Average and VoteCount are the
// plain (legacy) aggregates; WeightedAverage and WeightSum use the recency
// weight from database.VoteWeightSQL, where a vote's weight halves for every
// full year since it was cast.
type AverageRating struct {
	ClassProfessorID uint    `gorm:"column:class_professor_id"`
	Average          float64 `gorm:"column:AVG(nota)"`
	VoteCount        int64   `gorm:"column:COUNT(*)"`
	WeightedAverage  float64 `gorm:"column:weighted_avg"`
	WeightSum        float64 `gorm:"column:weight_sum"`
}

func (AverageRating) TableName() string {
	return "ListaMedias"
}
