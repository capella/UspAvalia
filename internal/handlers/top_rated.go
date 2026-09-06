package handlers

import (
	"fmt"
	"net/http"
	"uspavalia/internal/database"
	"uspavalia/internal/models"

	"github.com/sirupsen/logrus"
)

// topRatedLimit is how many rows each list on /destaques shows.
const topRatedLimit = 10

// BestRatedProfessor is a row of the top-rated professors list.
type BestRatedProfessor struct {
	ProfessorID   uint    `json:"professor_id"`
	ProfessorName string  `json:"professor_name"`
	UnitName      string  `json:"unit_name"`
	Average       float64 `json:"average"`
	VoteCount     int     `json:"vote_count"`
}

// BestRatedUnit is a row of the top-rated units list.
type BestRatedUnit struct {
	UnitID    uint    `json:"unit_id"`
	UnitName  string  `json:"unit_name"`
	Average   float64 `json:"average"`
	VoteCount int     `json:"vote_count"`
}

// weightedAverageSelect returns the SELECT fragment shared by the top-rated
// queries: the recency-weighted average on the 0-10 scale, the raw vote
// count and the total weight (used as a tie breaker so recent votes win).
// The votes table must be aliased as "v".
func (s *Server) weightedAverageSelect() string {
	weight := database.VoteWeightSQL(database.Dialect(s.db), "v.time")
	return fmt.Sprintf(`
			(SUM(v.score * %s) / SUM(%s)) * 2 AS average,
			COUNT(*) AS vote_count,
			SUM(%s) AS weight_sum`, weight, weight, weight)
}

// bestRatedProfessors ranks professors by the recency-weighted average of
// all their votes (any class), requiring at least MinVotesForTopRated votes.
func (s *Server) bestRatedProfessors(limit int) ([]BestRatedProfessor, error) {
	var rows []BestRatedProfessor
	err := s.db.Raw(`
		SELECT
			p.id AS professor_id,
			p.name AS professor_name,
			u.name AS unit_name,`+s.weightedAverageSelect()+`
		FROM votes v
		INNER JOIN class_professors ap ON v.class_professor_id = ap.id
		INNER JOIN professors p ON ap.professor_id = p.id
		INNER JOIN units u ON p.unit_id = u.id
		WHERE v.type <> ?
		GROUP BY p.id, p.name, u.name
		HAVING COUNT(*) >= ?
		ORDER BY average DESC, weight_sum DESC
		LIMIT ?
	`, models.VoteTypeDifficulty, models.MinVotesForTopRated, limit).Scan(&rows).Error
	return rows, err
}

// bestRatedUnits ranks academic units by the recency-weighted average of the
// votes on their disciplines, requiring at least MinVotesForTopRated votes.
func (s *Server) bestRatedUnits(limit int) ([]BestRatedUnit, error) {
	var rows []BestRatedUnit
	err := s.db.Raw(`
		SELECT
			u.id AS unit_id,
			u.name AS unit_name,`+s.weightedAverageSelect()+`
		FROM votes v
		INNER JOIN class_professors ap ON v.class_professor_id = ap.id
		INNER JOIN disciplines d ON ap.class_id = d.id
		INNER JOIN units u ON d.unit_id = u.id
		WHERE v.type <> ?
		GROUP BY u.id, u.name
		HAVING COUNT(*) >= ?
		ORDER BY average DESC, weight_sum DESC
		LIMIT ?
	`, models.VoteTypeDifficulty, models.MinVotesForTopRated, limit).Scan(&rows).Error
	return rows, err
}

// handleTopRated renders /destaques: the best rated class-professors (from
// the Melhores view), professors and units. Every list uses the recency
// weight from database.VoteWeightSQL, so recent votes take priority.
func (s *Server) handleTopRated(w http.ResponseWriter, r *http.Request) {
	var bestRatedDisciplines []models.BestRated
	if err := s.db.Limit(topRatedLimit).Find(&bestRatedDisciplines).Error; err != nil {
		logrus.Printf("Warning: Could not load best rated disciplines: %v", err)
	}

	bestRatedProfessors, err := s.bestRatedProfessors(topRatedLimit)
	if err != nil {
		logrus.Printf("Warning: Could not load best rated professors: %v", err)
	}

	bestRatedUnits, err := s.bestRatedUnits(topRatedLimit)
	if err != nil {
		logrus.Printf("Warning: Could not load best rated units: %v", err)
	}

	data := PageData{
		User: s.getCurrentUser(r),
		Data: map[string]interface{}{
			"BestRatedDisciplines": bestRatedDisciplines,
			"BestRatedProfessors":  bestRatedProfessors,
			"BestRatedUnits":       bestRatedUnits,
			"MinVotes":             models.MinVotesForTopRated,
		},
	}

	s.renderTemplate(w, r, "10melhores", data)
}
