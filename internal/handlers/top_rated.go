package handlers

import (
	"fmt"
	"net/http"
	"sync"
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

// topRatedData is everything the top-rated lists need, computed at once.
type topRatedData struct {
	Disciplines []models.BestRated
	Professors  []BestRatedProfessor
	Units       []BestRatedUnit
}

// topRatedCache memoizes topRatedData. The three ranking queries aggregate
// the whole votes table, and the result only drifts as votes come in (or as
// they age past a yearly boundary), so serving it for a while is fine.
type topRatedCache struct {
	mu         sync.RWMutex
	data       *topRatedData
	expiration time.Time
}

const topRatedCacheDuration = time.Hour

func (c *topRatedCache) get() *topRatedData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.data != nil && time.Now().Before(c.expiration) {
		return c.data
	}
	return nil
}

func (c *topRatedCache) set(data *topRatedData, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	c.expiration = time.Now().Add(duration)
}

// loadTopRated returns the top-rated lists, from cache when fresh. Lists
// that fail to load are logged and left empty; the result is still cached
// so a failing query does not get hammered.
func (s *Server) loadTopRated() *topRatedData {
	if cached := s.topRated.get(); cached != nil {
		return cached
	}

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

	s.topRated.set(data, topRatedCacheDuration)
	return data
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
