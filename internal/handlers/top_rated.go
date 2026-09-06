package handlers

import (
	"fmt"
	"net/http"
	"time"
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
	RecentVotes   int     `json:"recent_votes"`
	Ranking       float64 `json:"ranking"`
}

// BestRatedUnit is a row of the top-rated units list.
type BestRatedUnit struct {
	UnitID      uint    `json:"unit_id"`
	UnitName    string  `json:"unit_name"`
	Average     float64 `json:"average"`
	VoteCount   int     `json:"vote_count"`
	RecentVotes int     `json:"recent_votes"`
	Ranking     float64 `json:"ranking"`
}

// weightedAverageSelect returns the SELECT fragment shared by the top-rated
// queries: the recency-weighted average on the 0-10 scale, the raw vote
// count, the number of votes in the last year, the total weight and the
// ranking score the lists are ordered by (see database.RankingSQL). The
// votes table must be aliased as "v".
func (s *Server) weightedAverageSelect() string {
	dialect := database.Dialect(s.db)
	weight := database.VoteWeightSQL(dialect, "v.time")
	weightedSum := fmt.Sprintf("SUM(v.score * %s)", weight)
	weightSum := fmt.Sprintf("SUM(%s)", weight)
	return fmt.Sprintf(`
			(%s / %s) * 2 AS average,
			COUNT(*) AS vote_count,
			SUM(%s) AS recent_votes,
			%s AS weight_sum,
			%s AS ranking`,
		weightedSum, weightSum,
		database.RecentVoteSQL(dialect, "v.time"),
		weightSum,
		database.RankingSQL(dialect, weightedSum, weightSum))
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
		ORDER BY ranking DESC, weight_sum DESC
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
		ORDER BY ranking DESC, weight_sum DESC
		LIMIT ?
	`, models.VoteTypeDifficulty, models.MinVotesForTopRated, limit).Scan(&rows).Error
	return rows, err
}

// topRatedData is everything the top-rated lists need, computed at once.
type topRatedData struct {
	Disciplines []models.BestRated
	Professors  []BestRatedProfessor
	Units       []BestRatedUnit
}

// topRatedCacheDuration is how long the ranking lists are served from
// memory. The three ranking queries aggregate the whole votes table, and
// the result only drifts as votes come in (or as they age past a yearly
// boundary), so serving it for a while is fine.
const topRatedCacheDuration = time.Hour

// loadTopRated returns the top-rated lists, from cache when fresh. Lists
// that fail to load are logged and left empty; the result is still cached
// so a failing query does not get hammered.
func (s *Server) loadTopRated() *topRatedData {
	return s.topRatedCache.GetOrLoad(topRatedCacheDuration, func() *topRatedData {
		data := &topRatedData{}
		if err := s.db.Limit(topRatedLimit).Find(&data.Disciplines).Error; err != nil {
			logrus.Printf("Warning: Could not load best rated disciplines: %v", err)
		}

		var err error
		if data.Professors, err = s.bestRatedProfessors(topRatedLimit); err != nil {
			logrus.Printf("Warning: Could not load best rated professors: %v", err)
		}
		if data.Units, err = s.bestRatedUnits(topRatedLimit); err != nil {
			logrus.Printf("Warning: Could not load best rated units: %v", err)
		}
		return data
	})
}

// handleTopRated renders /destaques: the best rated class-professors (from
// the Melhores view), professors and units. Every list uses the recency
// weight from database.VoteWeightSQL, so recent votes take priority.
func (s *Server) handleTopRated(w http.ResponseWriter, r *http.Request) {
	topRated := s.loadTopRated()

	data := PageData{
		User: s.getCurrentUser(r),
		Data: map[string]interface{}{
			"BestRatedDisciplines": topRated.Disciplines,
			"BestRatedProfessors":  topRated.Professors,
			"BestRatedUnits":       topRated.Units,
			"MinVotes":             models.MinVotesForTopRated,
		},
	}

	s.renderTemplate(w, r, "10melhores", data)
}
