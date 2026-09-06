package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"uspavalia/internal/database"
	"uspavalia/internal/models"

	"github.com/sirupsen/logrus"
)

// topRatedLimit is how many rows each list on /destaques shows.
const topRatedLimit = 10

// BestRatedProfessor is a row of the top-rated professors list.
type BestRatedProfessor struct {
	ProfessorID      uint            `json:"professor_id"`
	ProfessorName    string          `json:"professor_name"`
	UnitName         string          `json:"unit_name"`
	Average          float64         `json:"average"`
	VoteCount        int             `json:"vote_count"`
	Evaluators       int             `json:"evaluators"`
	RecentEvaluators int             `json:"recent_evaluators"`
	Movement         models.Movement `json:"-"                 gorm:"-"`
}

// BestRatedUnit is a row of the top-rated units list.
type BestRatedUnit struct {
	UnitID           uint            `json:"unit_id"`
	UnitName         string          `json:"unit_name"`
	Average          float64         `json:"average"`
	VoteCount        int             `json:"vote_count"`
	Evaluators       int             `json:"evaluators"`
	RecentEvaluators int             `json:"recent_evaluators"`
	Movement         models.Movement `json:"-"                 gorm:"-"`
}

// rankingSelect returns the SELECT fragment shared by the top-rated
// queries: the recency-weighted average on the 0-10 scale, the raw vote
// count, the distinct evaluators overall and in the last year, and the
// total weight, all measured as of the unix timestamp at. The votes table
// must be aliased as "v"; topRatedOrder sorts these columns.
func (s *Server) rankingSelect(at int64) string {
	dialect := database.Dialect(s.db)
	now := strconv.FormatInt(at, 10)
	weight := database.VoteWeightSQLAt(dialect, "v.time", now)
	return fmt.Sprintf(`
			(SUM(v.score * %s) / SUM(%s)) * 2 AS average,
			COUNT(*) AS vote_count,
			COUNT(DISTINCT v.user_id) AS evaluators,
			COUNT(DISTINCT %s) AS recent_evaluators,
			SUM(%s) AS weight_sum`,
		weight, weight, database.RecentEvaluatorSQLAt("v.time", "v.user_id", now), weight)
}

// topRatedOrder is the ORDER BY clause matching rankingSelect.
var topRatedOrder = database.TopRatedOrderSQL("recent_evaluators", "average", "weight_sum")

// rankingQuery builds a top-rated query as of the unix timestamp at, using
// only votes cast before then. columns and joins name the entity (grouped
// by groupBy); the votes table is aliased "v" and class_professors "ap".
func (s *Server) rankingQuery(at int64, columns, joins, groupBy string, limit int, dest interface{}) error {
	return s.db.Raw(`
		SELECT `+columns+`,`+s.rankingSelect(at)+`
		FROM votes v
		INNER JOIN class_professors ap ON v.class_professor_id = ap.id
		`+joins+`
		WHERE v.type <> ? AND v.time <= ?
		GROUP BY `+groupBy+`
		HAVING COUNT(DISTINCT v.user_id) >= ?
		`+topRatedOrder+`
		LIMIT ?
	`, models.VoteTypeDifficulty, at, models.MinEvaluatorsForTopRated, limit).Scan(dest).Error
}

// bestRatedProfessors ranks professors by the recency-weighted average of
// all their votes (any class), requiring MinEvaluatorsForTopRated evaluators.
func (s *Server) bestRatedProfessors(limit int) ([]BestRatedProfessor, error) {
	return s.bestRatedProfessorsAt(limit, time.Now().Unix())
}

func (s *Server) bestRatedProfessorsAt(limit int, at int64) ([]BestRatedProfessor, error) {
	var rows []BestRatedProfessor
	err := s.rankingQuery(at,
		"p.id AS professor_id, p.name AS professor_name, u.name AS unit_name",
		`INNER JOIN professors p ON ap.professor_id = p.id
		INNER JOIN units u ON p.unit_id = u.id`,
		"p.id, p.name, u.name", limit, &rows)
	return rows, err
}

// bestRatedUnits ranks academic units by the recency-weighted average of the
// votes on their disciplines, requiring MinEvaluatorsForTopRated evaluators.
func (s *Server) bestRatedUnits(limit int) ([]BestRatedUnit, error) {
	return s.bestRatedUnitsAt(limit, time.Now().Unix())
}

func (s *Server) bestRatedUnitsAt(limit int, at int64) ([]BestRatedUnit, error) {
	var rows []BestRatedUnit
	err := s.rankingQuery(at,
		"u.id AS unit_id, u.name AS unit_name",
		`INNER JOIN disciplines d ON ap.class_id = d.id
		INNER JOIN units u ON d.unit_id = u.id`,
		"u.id, u.name", limit, &rows)
	return rows, err
}

// bestRatedDisciplineIDsAt returns the class-professor ids the Melhores view
// would list as of the unix timestamp at, in order.
func (s *Server) bestRatedDisciplineIDsAt(limit int, at int64) ([]uint, error) {
	var rows []struct{ ID uint }
	err := s.rankingQuery(at, "ap.id AS id", "", "ap.id", limit, &rows)
	ids := make([]uint, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	return ids, err
}

// topRatedData is everything the top-rated lists need, computed at once.
type topRatedData struct {
	Disciplines []models.BestRated
	Professors  []BestRatedProfessor
	Units       []BestRatedUnit
}

// topRatedCacheDuration is how long the ranking lists are served from
// memory. The ranking queries aggregate the whole votes table, and the
// result only drifts as votes come in (or as they age past a yearly
// boundary), so serving it for a while is fine.
const topRatedCacheDuration = time.Hour

// loadTopRated returns the top-rated lists, from cache when fresh. Lists
// that fail to load are logged and left empty; the result is still cached
// so a failing query does not get hammered. Each row is annotated with its
// movement since the end of the previous semester.
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

		s.annotateMovement(data, semesterStart(time.Now()).Unix())
		return data
	})
}

// semesterStart returns the first day of the USP semester containing t:
// February 1 for the first semester (February to July) and August 1 for
// the second (August to January).
func semesterStart(t time.Time) time.Time {
	y, m := t.Year(), t.Month()
	switch {
	case m == time.January:
		return time.Date(y-1, time.August, 1, 0, 0, 0, 0, t.Location())
	case m < time.August:
		return time.Date(y, time.February, 1, 0, 0, 0, 0, t.Location())
	default:
		return time.Date(y, time.August, 1, 0, 0, 0, 0, t.Location())
	}
}

// annotateMovement fills in the Movement of every row by re-ranking each
// list as of baseline (a unix timestamp, the start of the current semester)
// using only the votes cast before it. Votes are never edited, so that is
// exactly the list as it stood at the end of the previous semester, and no
// snapshot needs to be stored. A list with no votes before the baseline
// gets no movement information.
func (s *Server) annotateMovement(data *topRatedData, baseline int64) {
	ids, err := s.bestRatedDisciplineIDsAt(topRatedLimit, baseline)
	if err != nil {
		logrus.Printf("Warning: Could not rank disciplines as of %d: %v", baseline, err)
	}
	for i, m := range movements(ids, len(data.Disciplines), func(i int) uint { return data.Disciplines[i].ID }) {
		data.Disciplines[i].Movement = m
	}

	profs, err := s.bestRatedProfessorsAt(topRatedLimit, baseline)
	if err != nil {
		logrus.Printf("Warning: Could not rank professors as of %d: %v", baseline, err)
	}
	ids = ids[:0]
	for _, p := range profs {
		ids = append(ids, p.ProfessorID)
	}
	for i, m := range movements(ids, len(data.Professors), func(i int) uint { return data.Professors[i].ProfessorID }) {
		data.Professors[i].Movement = m
	}

	units, err := s.bestRatedUnitsAt(topRatedLimit, baseline)
	if err != nil {
		logrus.Printf("Warning: Could not rank units as of %d: %v", baseline, err)
	}
	ids = ids[:0]
	for _, u := range units {
		ids = append(ids, u.UnitID)
	}
	for i, m := range movements(ids, len(data.Units), func(i int) uint { return data.Units[i].UnitID }) {
		data.Units[i].Movement = m
	}
}

// movements compares the current order (n entries, id(i) for the i-th) with
// the baseline order and returns one Movement per current entry. An empty
// baseline yields unknown movements.
func movements(baseline []uint, n int, id func(int) uint) []models.Movement {
	moves := make([]models.Movement, n)
	if len(baseline) == 0 {
		return moves
	}
	positions := make(map[uint]int, len(baseline))
	for i, b := range baseline {
		positions[b] = i + 1
	}
	for i := range moves {
		pos, ok := positions[id(i)]
		moves[i] = models.Movement{Known: true, New: !ok}
		if ok {
			moves[i].Delta = pos - (i + 1)
		}
	}
	return moves
}

// handleTopRated renders /destaques: the best rated class-professors (from
// the Melhores view), professors and units. Every list uses the recency
// weight from database.VoteWeightSQL and the order of
// database.TopRatedOrderSQL, so recent evaluations take priority.
func (s *Server) handleTopRated(w http.ResponseWriter, r *http.Request) {
	topRated := s.loadTopRated()

	data := PageData{
		User: s.getCurrentUser(r),
		Data: map[string]interface{}{
			"BestRatedDisciplines": topRated.Disciplines,
			"BestRatedProfessors":  topRated.Professors,
			"BestRatedUnits":       topRated.Units,
			"MinEvaluators":        models.MinEvaluatorsForTopRated,
			"MinRecentEvaluators":  models.MinRecentEvaluatorsForTopRated,
		},
	}

	s.renderTemplate(w, r, "10melhores", data)
}
