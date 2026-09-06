package models

// MinVotesForTopRated is the minimum number of votes (excluding the
// difficulty vote type) a class-professor, professor or unit needs to appear
// in the top-rated lists (/destaques).
const MinVotesForTopRated = 15

// TopRatedPriorWeight is the weight of the prior in the top-rated ranking,
// measured in recent votes (a vote cast in the last year weighs 1). An entry
// needs about this much recent weight before its own average dominates the
// global average in the ranking; two evaluations of four votes each is
// enough. See database.RankingSQL.
const TopRatedPriorWeight = 8

// AverageRating maps the ListaMedias view. Average and VoteCount are the
// plain (legacy) aggregates; WeightedAverage and WeightSum use the recency
// weight from database.VoteWeightSQL, where a vote's weight halves for every
// full year since it was cast. RecentVotes counts votes from the last year
// and Ranking is the score the top-rated lists are ordered by.
type AverageRating struct {
	ClassProfessorID uint    `gorm:"column:class_professor_id"`
	Average          float64 `gorm:"column:AVG(nota)"`
	VoteCount        int64   `gorm:"column:COUNT(*)"`
	WeightedAverage  float64 `gorm:"column:weighted_avg"`
	WeightSum        float64 `gorm:"column:weight_sum"`
	RecentVotes      int64   `gorm:"column:recent_votes"`
	Ranking          float64 `gorm:"column:ranking"`
}

func (AverageRating) TableName() string {
	return "ListaMedias"
}
