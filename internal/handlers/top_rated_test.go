package handlers

import (
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
	// Calling seedVotes twice with the same unit reuses the same rows.
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
	for i := 0; i < n; i++ {
		v := models.Vote{
			ClassProfessorID: cp.ID,
			UserID:           "user",
			Time:             voteTime(ageYears),
			Score:            score,
			Type:             int(models.VoteTypeGeneral),
		}
		if err := db.Create(&v).Error; err != nil {
			t.Fatalf("create vote: %v", err)
		}
	}
}

func TestBestRatedProfessorsAndUnitsWeightByRecency(t *testing.T) {
	s := newTopRatedServer(t)
	// Old (2.5 years) perfect scores weigh 1/4: 20 votes of 5 count like 5.
	// Together with 20 recent votes of 3 the weighted average is 3.4 (6.8/10).
	seedVotes(t, s.db, "Old", 5, 20, 2.5)
	seedVotes(t, s.db, "Old", 3, 20, 0.1)
	// Recent, consistently 4 (8/10) with far fewer votes.
	seedVotes(t, s.db, "Recent", 4, 15, 0.1)
	// Below the vote threshold: must not appear.
	seedVotes(t, s.db, "Few", 5, models.MinVotesForTopRated-1, 0.1)

	profs, err := s.bestRatedProfessors(10)
	if err != nil {
		t.Fatalf("bestRatedProfessors: %v", err)
	}
	if len(profs) != 2 {
		t.Fatalf("professors = %d, want 2", len(profs))
	}
	if profs[0].ProfessorName != "Professor Recent" || profs[1].ProfessorName != "Professor Old" {
		t.Errorf("professor order = [%s %s], want [Professor Recent Professor Old]", profs[0].ProfessorName, profs[1].ProfessorName)
	}
	if math.Abs(profs[0].Average-8) > 1e-9 || math.Abs(profs[1].Average-6.8) > 1e-9 {
		t.Errorf("professor averages = [%v %v], want [8 6.8]", profs[0].Average, profs[1].Average)
	}
	if profs[1].VoteCount != 40 {
		t.Errorf("Professor Old vote count = %d, want raw count 40", profs[1].VoteCount)
	}

	units, err := s.bestRatedUnits(10)
	if err != nil {
		t.Fatalf("bestRatedUnits: %v", err)
	}
	if len(units) != 2 {
		t.Fatalf("units = %d, want 2", len(units))
	}
	if units[0].UnitName != "Recent" || units[1].UnitName != "Old" {
		t.Errorf("unit order = [%s %s], want [Recent Old]", units[0].UnitName, units[1].UnitName)
	}
	if math.Abs(units[0].Average-8) > 1e-9 || math.Abs(units[1].Average-6.8) > 1e-9 {
		t.Errorf("unit averages = [%v %v], want [8 6.8]", units[0].Average, units[1].Average)
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
	seedVotes(t, s.db, "Escola Politécnica", 4, 15, 0.1)

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
		"Escola Politécnica",
		"Professor Escola Politécnica",
		"cai pela metade a cada ano",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page does not contain %q", want)
		}
	}
	if got := strings.Count(body, "8.00"); got < 3 {
		t.Errorf("expected the 8.00 average in all three lists, found it %d times", got)
	}
}
