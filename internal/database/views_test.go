package database

import (
	"math"
	"strings"
	"testing"
	"time"
	"uspavalia/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := autoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := CreateViews(db); err != nil {
		t.Fatalf("create views: %v", err)
	}
	return db
}

// seedClassProfessor creates a unit, discipline, professor and the
// class-professor row joining them, returning the class-professor id.
func seedClassProfessor(t *testing.T, db *gorm.DB, code string) uint {
	t.Helper()
	unit := models.Unit{Name: "Unit " + code}
	if err := db.Create(&unit).Error; err != nil {
		t.Fatalf("create unit: %v", err)
	}
	disc := models.Discipline{Name: "Discipline " + code, Code: code, UnitID: unit.ID}
	if err := db.Create(&disc).Error; err != nil {
		t.Fatalf("create discipline: %v", err)
	}
	prof := models.Professor{Name: "Professor " + code, UnitID: unit.ID}
	if err := db.Create(&prof).Error; err != nil {
		t.Fatalf("create professor: %v", err)
	}
	cp := models.ClassProfessor{ClassID: disc.ID, ProfessorID: prof.ID}
	if err := db.Create(&cp).Error; err != nil {
		t.Fatalf("create class professor: %v", err)
	}
	return cp.ID
}

// voteTime returns a unix timestamp ageYears in the past.
func voteTime(ageYears float64) int64 {
	return time.Now().Unix() - int64(ageYears*secondsPerYear)
}

func addVotes(t *testing.T, db *gorm.DB, cpID uint, score, n int, ageYears float64) {
	t.Helper()
	for i := 0; i < n; i++ {
		v := models.Vote{
			ClassProfessorID: cpID,
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

func TestVoteWeightSQLHalvesEveryYear(t *testing.T) {
	db := newTestDB(t)
	expr := VoteWeightSQL("sqlite", "?")

	cases := []struct {
		ageYears float64
		want     float64
	}{
		{-1, 1},    // future timestamp is clamped to full weight
		{0, 1},     // just cast
		{0.5, 1},   // less than a year old
		{1.5, 0.5}, // one full year
		{2.5, 0.25},
		{3.5, 0.125},
		{10.5, 1.0 / 1024},
	}
	for _, c := range cases {
		var got float64
		if err := db.Raw("SELECT "+expr, voteTime(c.ageYears)).Scan(&got).Error; err != nil {
			t.Fatalf("age %v: %v", c.ageYears, err)
		}
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("weight for a vote %v years old = %v, want %v", c.ageYears, got, c.want)
		}
	}
}

func TestVoteWeightSQLMySQL(t *testing.T) {
	expr := VoteWeightSQL("mysql", "v.time")
	// 1E0 (a DOUBLE) matters: with the DECIMAL literal 1.0 MySQL rounds the
	// weights to five decimals and old all-5 votes averaged above 10.
	for _, want := range []string{"UNIX_TIMESTAMP()", "v.time", "GREATEST(0, LEAST(62", "1E0 / (1 <<"} {
		if !strings.Contains(expr, want) {
			t.Errorf("mysql expression %q does not contain %q", expr, want)
		}
	}
}

func TestListaMediasWeightedAverage(t *testing.T) {
	db := newTestDB(t)
	cp := seedClassProfessor(t, db, "A")
	addVotes(t, db, cp, 5, 10, 2.5) // weight 0.25 each
	addVotes(t, db, cp, 1, 5, 0.1)  // weight 1 each

	var row models.AverageRating
	if err := db.First(&row).Error; err != nil {
		t.Fatalf("query ListaMedias: %v", err)
	}
	if row.VoteCount != 15 {
		t.Errorf("COUNT(*) = %d, want 15", row.VoteCount)
	}
	if math.Abs(row.Average-55.0/15) > 1e-9 {
		t.Errorf("AVG(nota) = %v, want %v", row.Average, 55.0/15)
	}
	// (10*5*0.25 + 5*1*1) / (10*0.25 + 5)
	wantWeighted := (12.5 + 5) / 7.5
	if math.Abs(row.WeightedAverage-wantWeighted) > 1e-9 {
		t.Errorf("weighted_avg = %v, want %v", row.WeightedAverage, wantWeighted)
	}
	if math.Abs(row.WeightSum-7.5) > 1e-9 {
		t.Errorf("weight_sum = %v, want 7.5", row.WeightSum)
	}
}

func TestMelhoresPrioritizesRecentVotes(t *testing.T) {
	db := newTestDB(t)

	// A: great three years ago, mediocre now. Plain average 4.0.
	cpA := seedClassProfessor(t, db, "A")
	addVotes(t, db, cpA, 5, 15, 3.5)
	addVotes(t, db, cpA, 3, 15, 0.1)

	// B: consistently good now. Plain average 4.0 as well.
	cpB := seedClassProfessor(t, db, "B")
	addVotes(t, db, cpB, 4, 15, 0.1)

	// C: perfect but below the vote threshold.
	cpC := seedClassProfessor(t, db, "C")
	addVotes(t, db, cpC, 5, models.MinVotesForTopRated-1, 0.1)

	var rows []models.BestRated
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("query Melhores: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("Melhores rows = %d, want 2 (C has too few votes)", len(rows))
	}
	if rows[0].ID != cpB || rows[1].ID != cpA {
		t.Fatalf("Melhores order = [%d %d], want [%d %d]", rows[0].ID, rows[1].ID, cpB, cpA)
	}
	if math.Abs(rows[0].Average-8.0) > 1e-9 {
		t.Errorf("B media = %v, want 8", rows[0].Average)
	}
	// A: (15*5*0.125 + 15*3) / (15*0.125 + 15) * 2
	wantA := (9.375 + 45) / 16.875 * 2
	if math.Abs(rows[1].Average-wantA) > 1e-9 {
		t.Errorf("A media = %v, want %v", rows[1].Average, wantA)
	}
	if rows[1].VoteCount != 30 {
		t.Errorf("A votos = %d, want the raw count 30", rows[1].VoteCount)
	}
	if rows[0].RecentVotes != 15 || rows[1].RecentVotes != 15 {
		t.Errorf("votos_recentes = [%d %d], want [15 15]", rows[0].RecentVotes, rows[1].RecentVotes)
	}
}

// A perfect average backed only by old votes must not stay on top forever:
// entries with enough votes in the last year come first, whatever their
// average, and the stale 10 is listed after them.
func TestMelhoresDemotesStaleTens(t *testing.T) {
	db := newTestDB(t)

	// A: 10.00 from 16 votes three years ago, nothing this year.
	cpA := seedClassProfessor(t, db, "A")
	addVotes(t, db, cpA, 5, 16, 3.5)

	// B: 9.00 from 40 votes this year.
	cpB := seedClassProfessor(t, db, "B")
	addVotes(t, db, cpB, 5, 20, 0.1)
	addVotes(t, db, cpB, 4, 20, 0.1)

	// C: 4.00 from 30 votes this year.
	cpC := seedClassProfessor(t, db, "C")
	addVotes(t, db, cpC, 2, 30, 0.1)

	// D: 10.00 but only one evaluation (4 votes) this year: not enough.
	cpD := seedClassProfessor(t, db, "D")
	addVotes(t, db, cpD, 5, 12, 2.5)
	addVotes(t, db, cpD, 5, 4, 0.1)

	var rows []models.BestRated
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("query Melhores: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("Melhores rows = %d, want 4", len(rows))
	}
	got := []uint{rows[0].ID, rows[1].ID, rows[2].ID, rows[3].ID}
	want := []uint{cpB, cpC, cpD, cpA}
	if got[0] != want[0] || got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
		t.Fatalf("Melhores order = %v, want %v (B, C active; then D, A by average and weight)", got, want)
	}
	// The displayed average is still the weighted average.
	if math.Abs(rows[0].Average-9) > 1e-9 || math.Abs(rows[3].Average-10) > 1e-9 {
		t.Errorf("averages = [%v ... %v], want [9 ... 10]", rows[0].Average, rows[3].Average)
	}
	if rows[0].RecentVotes != 40 || rows[2].RecentVotes != 4 || rows[3].RecentVotes != 0 {
		t.Errorf("votos_recentes = [%d %d %d %d], want [40 30 4 0]",
			rows[0].RecentVotes, rows[1].RecentVotes, rows[2].RecentVotes, rows[3].RecentVotes)
	}
}

// Among entries with enough recent votes the average decides: a 10 backed by
// two recent evaluations beats a 9.9 backed by many more recent votes.
func TestMelhoresOrdersActiveEntriesByAverage(t *testing.T) {
	db := newTestDB(t)

	cpTen := seedClassProfessor(t, db, "Ten")
	addVotes(t, db, cpTen, 5, 24, 4.5) // old but perfect
	addVotes(t, db, cpTen, 5, 8, 0.1)  // exactly MinRecentVotesForTopRated

	cpNine := seedClassProfessor(t, db, "Nine")
	addVotes(t, db, cpNine, 5, 50, 0.1)
	addVotes(t, db, cpNine, 4, 2, 0.1)

	var rows []models.BestRated
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("query Melhores: %v", err)
	}
	if len(rows) != 2 || rows[0].ID != cpTen || rows[1].ID != cpNine {
		t.Fatalf("Melhores order = %+v, want Ten then Nine", rows)
	}
	if math.Abs(rows[0].Average-10) > 1e-9 || rows[0].RecentVotes != 8 {
		t.Errorf("Ten = %.2f with %d recent votes, want 10.00 with 8", rows[0].Average, rows[0].RecentVotes)
	}
}
