package handlers

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
	"uspavalia/internal/database"
	"uspavalia/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTopRatedServer(t *testing.T) *Server {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Unit{},
		&models.Discipline{},
		&models.Professor{},
		&models.ClassProfessor{},
		&models.Vote{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.CreateViews(db); err != nil {
		t.Fatalf("create views: %v", err)
	}
	return &Server{db: db}
}

func voteTime(ageYears float64) int64 {
	return time.Now().Unix() - int64(ageYears*365.25*24*3600)
}

// seedVotes creates a unit, a discipline and a professor named after the
// unit, joins them and casts n votes with the given score and age.
func seedVotes(t *testing.T, db *gorm.DB, unitName string, score, n int, ageYears float64) {
	t.Helper()
	seedVotesAt(t, db, unitName, score, n, voteTime(ageYears))
}

var userSeq int

// seedVotesAt is seedVotes with an explicit unix timestamp. Like a real
// evaluation, every four votes come from one distinct evaluator. Calling it
// twice with the same unit reuses the same rows.
func seedVotesAt(t *testing.T, db *gorm.DB, unitName string, score, n int, ts int64) {
	t.Helper()
	var unit models.Unit
	if err := db.FirstOrCreate(&unit, models.Unit{Name: unitName}).Error; err != nil {
		t.Fatalf("create unit: %v", err)
	}
	var disc models.Discipline
	if err := db.FirstOrCreate(&disc, models.Discipline{Name: "Discipline " + unitName, Code: "CODE-" + unitName, UnitID: unit.ID}).Error; err != nil {
		t.Fatalf("create discipline: %v", err)
	}
	var prof models.Professor
	if err := db.FirstOrCreate(&prof, models.Professor{Name: "Professor " + unitName, UnitID: unit.ID}).Error; err != nil {
		t.Fatalf("create professor: %v", err)
	}
	var cp models.ClassProfessor
	if err := db.FirstOrCreate(&cp, models.ClassProfessor{ClassID: disc.ID, ProfessorID: prof.ID}).Error; err != nil {
		t.Fatalf("create class professor: %v", err)
	}
	base := userSeq
	userSeq += (n + 3) / 4
	for i := 0; i < n; i++ {
		v := models.Vote{
			ClassProfessorID: cp.ID,
			UserID:           fmt.Sprintf("user%d", base+i/4),
			Time:             ts,
			Score:            score,
			Type:             int(models.VoteTypeGeneral),
		}
		if err := db.Create(&v).Error; err != nil {
			t.Fatalf("create vote: %v", err)
		}
	}
}

func professorNames(profs []BestRatedProfessor) string {
	var names []string
	for _, p := range profs {
		names = append(names, strings.TrimPrefix(p.ProfessorName, "Professor "))
	}
	return strings.Join(names, ",")
}

func unitNames(units []BestRatedUnit) string {
	var names []string
	for _, u := range units {
		names = append(names, u.UnitName)
	}
	return strings.Join(names, ",")
}

func TestBestRatedProfessorsAndUnitsWeightByRecency(t *testing.T) {
	s := newTopRatedServer(t)
	// Old (2.5 years) perfect scores weigh 1/4: 20 votes of 5 count like 5.
	// Together with 20 recent votes of 3 the weighted average is 3.4 (6.8/10).
	seedVotes(t, s.db, "Old", 5, 20, 2.5)
	seedVotes(t, s.db, "Old", 3, 20, 0.1)
	// Recent, consistently 4 (8/10) with fewer evaluators.
	seedVotes(t, s.db, "Recent", 4, 15, 0.1)
	// Three evaluators only: below the threshold, must not appear.
	seedVotes(t, s.db, "Few", 5, 4*(models.MinEvaluatorsForTopRated-1), 0.1)

	profs, err := s.bestRatedProfessors(10)
	if err != nil {
		t.Fatalf("bestRatedProfessors: %v", err)
	}
	if got := professorNames(profs); got != "Recent,Old" {
		t.Fatalf("professor order = %s, want Recent,Old", got)
	}
	if math.Abs(profs[0].Average-8) > 1e-9 || math.Abs(profs[1].Average-6.8) > 1e-9 {
		t.Errorf("professor averages = [%v %v], want [8 6.8]", profs[0].Average, profs[1].Average)
	}
	if profs[1].VoteCount != 40 || profs[1].Evaluators != 10 {
		t.Errorf("Professor Old votes = %d, evaluators = %d; want 40 and 10", profs[1].VoteCount, profs[1].Evaluators)
	}
	if profs[0].RecentEvaluators != 4 || profs[1].RecentEvaluators != 5 {
		t.Errorf("recent evaluators = [%d %d], want [4 5]", profs[0].RecentEvaluators, profs[1].RecentEvaluators)
	}

	units, err := s.bestRatedUnits(10)
	if err != nil {
		t.Fatalf("bestRatedUnits: %v", err)
	}
	if got := unitNames(units); got != "Recent,Old" {
		t.Fatalf("unit order = %s, want Recent,Old", got)
	}
	if math.Abs(units[0].Average-8) > 1e-9 || math.Abs(units[1].Average-6.8) > 1e-9 {
		t.Errorf("unit averages = [%v %v], want [8 6.8]", units[0].Average, units[1].Average)
	}
}

// Professors with enough evaluators this year come first, ordered by
// average, so a 10 backed by three recent evaluators beats a 9.9 with
// thirteen, and a 10 rated only years ago is listed after everyone active.
// Same for units.
func TestBestRatedPrioritizesRecentlyRated(t *testing.T) {
	s := newTopRatedServer(t)
	seedVotes(t, s.db, "Stale", 5, 16, 3.5)
	seedVotes(t, s.db, "Ten", 5, 24, 4.5)
	seedVotes(t, s.db, "Ten", 5, 12, 0.1)
	seedVotes(t, s.db, "Nine", 5, 50, 0.1)
	seedVotes(t, s.db, "Nine", 4, 2, 0.1)
	seedVotes(t, s.db, "Low", 2, 30, 0.1)

	profs, err := s.bestRatedProfessors(10)
	if err != nil {
		t.Fatalf("bestRatedProfessors: %v", err)
	}
	if got := professorNames(profs); got != "Ten,Nine,Low,Stale" {
		t.Fatalf("professor order = %s, want Ten,Nine,Low,Stale", got)
	}
	if math.Abs(profs[0].Average-10) > 1e-9 || profs[0].RecentEvaluators != 3 || profs[3].RecentEvaluators != 0 {
		t.Errorf("Ten = %.2f (%d recent), Stale recent = %d; want 10.00 (3), 0",
			profs[0].Average, profs[0].RecentEvaluators, profs[3].RecentEvaluators)
	}

	units, err := s.bestRatedUnits(10)
	if err != nil {
		t.Fatalf("bestRatedUnits: %v", err)
	}
	if got := unitNames(units); got != "Ten,Nine,Low,Stale" {
		t.Fatalf("unit order = %s, want Ten,Nine,Low,Stale", got)
	}
}

func TestSemesterStart(t *testing.T) {
	loc := time.UTC
	cases := map[time.Time]time.Time{
		time.Date(2026, time.January, 15, 0, 0, 0, 0, loc):  time.Date(2025, time.August, 1, 0, 0, 0, 0, loc),
		time.Date(2026, time.February, 1, 0, 0, 0, 0, loc):  time.Date(2026, time.February, 1, 0, 0, 0, 0, loc),
		time.Date(2026, time.July, 31, 23, 0, 0, 0, loc):    time.Date(2026, time.February, 1, 0, 0, 0, 0, loc),
		time.Date(2026, time.August, 1, 0, 0, 0, 0, loc):    time.Date(2026, time.August, 1, 0, 0, 0, 0, loc),
		time.Date(2026, time.December, 31, 0, 0, 0, 0, loc): time.Date(2026, time.August, 1, 0, 0, 0, 0, loc),
	}
	for in, want := range cases {
		if got := semesterStart(in); !got.Equal(want) {
			t.Errorf("semesterStart(%s) = %s, want %s", in.Format("2006-01-02"), got.Format("2006-01-02"), want.Format("2006-01-02"))
		}
	}
}

func TestMovements(t *testing.T) {
	ids := []uint{1, 2, 3}
	id := func(i int) uint { return ids[i] }

	for _, m := range movements(nil, 3, id) {
		if m.Known || m.Label() != "" {
			t.Errorf("without a baseline movement = %+v, want unknown", m)
		}
	}

	// Baseline order was 3, 1, 2; now it is 1, 2, 3.
	got := movements([]uint{3, 1, 2}, 3, id)
	want := []string{"▲1", "▲1", "▼2"}
	for i, m := range got {
		if m.Label() != want[i] {
			t.Errorf("movement[%d] = %q, want %q", i, m.Label(), want[i])
		}
	}
	if m := movements([]uint{3}, 3, id)[0]; !m.New || m.Label() != "novo" {
		t.Errorf("entry missing from the baseline = %+v, want novo", m)
	}
}

// The movement of each row compares the current order with the order as
// of the start of the current semester, computed from the votes cast
// before it: no snapshot table is involved.
func TestTopRatedMovementSinceLastSemester(t *testing.T) {
	s := newTopRatedServer(t)
	now := time.Now()
	start := semesterStart(now)
	before := start.Add(-30 * 24 * time.Hour).Unix()
	during := start.Unix() + (now.Unix()-start.Unix())/2 + 1

	// Last semester: A 10.0, B 8.0, C 6.0 (four evaluators each).
	seedVotesAt(t, s.db, "A", 5, 16, before)
	seedVotesAt(t, s.db, "B", 4, 16, before)
	seedVotesAt(t, s.db, "C", 3, 16, before)
	// This semester: C gets ten enthusiastic evaluators, D is new at 9.5.
	seedVotesAt(t, s.db, "C", 5, 40, during)
	seedVotesAt(t, s.db, "D", 5, 12, during)
	seedVotesAt(t, s.db, "D", 4, 4, during)

	data := s.loadTopRated()
	if got := professorNames(data.Professors); got != "A,D,C,B" {
		t.Fatalf("professor order = %s, want A,D,C,B", got)
	}
	want := []string{"", "novo", "", "▼2"} // A stays 1st, C stays 3rd, B fell from 2nd to 4th
	for i, p := range data.Professors {
		if !p.Movement.Known || p.Movement.Label() != want[i] {
			t.Errorf("professor %s movement = %q (known %v), want %q", p.ProfessorName, p.Movement.Label(), p.Movement.Known, want[i])
		}
	}
	if got := unitNames(data.Units); got != "A,D,C,B" {
		t.Fatalf("unit order = %s, want A,D,C,B", got)
	}
	for i, u := range data.Units {
		if u.Movement.Label() != want[i] {
			t.Errorf("unit %s movement = %q, want %q", u.UnitName, u.Movement.Label(), want[i])
		}
	}
	for i, d := range data.Disciplines {
		if d.Movement.Label() != want[i] {
			t.Errorf("discipline %s movement = %q, want %q", d.Code, d.Movement.Label(), want[i])
		}
	}
}

func TestLoadTopRatedIsCached(t *testing.T) {
	s := newTopRatedServer(t)
	seedVotes(t, s.db, "First", 4, 16, 0.1)

	first := s.loadTopRated()
	if len(first.Units) != 1 || len(first.Professors) != 1 || len(first.Disciplines) != 1 {
		t.Fatalf("first load = %d units, %d professors, %d disciplines; want 1 each",
			len(first.Units), len(first.Professors), len(first.Disciplines))
	}
	// Nothing was ranked before this semester started... unless the recent
	// votes fell before it; either way no movement label is shown.
	if first.Professors[0].Movement.Label() != "" {
		t.Errorf("movement = %q, want none", first.Professors[0].Movement.Label())
	}

	// New votes must not show up until the cache expires.
	seedVotes(t, s.db, "Second", 5, 16, 0.1)
	if again := s.loadTopRated(); again != first {
		t.Errorf("second load was not served from cache")
	}

	s.topRatedCache.Invalidate()
	if fresh := s.loadTopRated(); len(fresh.Units) != 2 {
		t.Errorf("units after cache expiry = %d, want 2", len(fresh.Units))
	}
}

func TestHandleTopRatedRendersThreeLists(t *testing.T) {
	// Templates are resolved relative to the repository root.
	wd, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	s := newTopRatedServer(t)
	seedVotes(t, s.db, "Escola Politécnica", 4, 16, 0.1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/destaques", nil)
	s.handleTopRated(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"Cursos Avaliados",
		"Professores Avaliados",
		"Unidades Avaliadas",
		"Avaliadores",
		"No último ano",
		"Evolução",
		"Escola Politécnica",
		"Professor Escola Politécnica",
		"cai pela metade a cada ano",
		"fim do semestre anterior",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page does not contain %q", want)
		}
	}
	if strings.Contains(body, "Alunos") {
		t.Errorf("page must say Avaliadores, not Alunos")
	}
	if got := strings.Count(body, "8.00"); got < 3 {
		t.Errorf("expected the 8.00 average in all three lists, found it %d times", got)
	}
}
