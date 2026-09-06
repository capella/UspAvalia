package handlers

import (
	"os"
	"testing"
	"uspavalia/internal/database"
	"uspavalia/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestTopRatedOnMySQL runs the ranking queries against a real MySQL, which
// is what production uses and where the DECIMAL rounding bug lived. It is
// skipped unless USPAVALIA_TEST_MYSQL_DSN points at a scratch database, e.g.
//
//	docker run --rm -d -p 3307:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=yes \
//	    -e MYSQL_DATABASE=uspavalia_test mysql
//	USPAVALIA_TEST_MYSQL_DSN='root:@tcp(127.0.0.1:3307)/uspavalia_test' \
//	    go test ./internal/handlers/ -run TestTopRatedOnMySQL -v
//
// The tables and views in that database are dropped and recreated.
func TestTopRatedOnMySQL(t *testing.T) {
	dsn := os.Getenv("USPAVALIA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("USPAVALIA_TEST_MYSQL_DSN not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	db.Exec("DROP VIEW IF EXISTS Melhores")
	db.Exec("DROP VIEW IF EXISTS ListaMedias")
	tables := []interface{}{&models.Vote{}, &models.ClassProfessor{}, &models.Professor{}, &models.Discipline{}, &models.Unit{}}
	if err := db.Migrator().DropTable(tables...); err != nil {
		t.Fatalf("drop tables: %v", err)
	}
	if err := db.AutoMigrate(tables...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.CreateViews(db); err != nil {
		t.Fatalf("create views: %v", err)
	}
	s := &Server{db: db}

	// All 5s, 12 years old, seven evaluators: the weights are 1/4096 and
	// the average must still be exactly 10 (it was 10.04 with DECIMAL
	// weights).
	seedVotes(t, db, "Antiga", 5, 28, 12.5)
	// All 5s this year from four evaluators: ranks first because it is
	// active.
	seedVotes(t, db, "Recente", 5, 16, 0.1)

	data := s.loadTopRated()
	profs := data.Professors
	if len(profs) != 2 {
		t.Fatalf("professors = %d, want 2", len(profs))
	}
	for _, p := range profs {
		if p.Average != 10 {
			t.Errorf("%s average = %v, want exactly 10", p.ProfessorName, p.Average)
		}
	}
	if profs[0].ProfessorName != "Professor Recente" || profs[0].RecentEvaluators != 4 || profs[1].RecentEvaluators != 0 {
		t.Errorf("order/recent evaluators = %+v, want Recente (4 recent) first", profs)
	}
	// Antiga was already ranked last semester, so a baseline exists.
	for _, p := range profs {
		if !p.Movement.Known {
			t.Errorf("%s movement unknown, want a baseline from the previous semester", p.ProfessorName)
		}
	}

	for _, u := range data.Units {
		if u.Average != 10 {
			t.Errorf("unit %s average = %v, want exactly 10", u.UnitName, u.Average)
		}
	}

	best := data.Disciplines
	if len(best) != 2 {
		t.Fatalf("Melhores rows = %d, want 2", len(best))
	}
	for _, b := range best {
		if b.Average != 10 {
			t.Errorf("Melhores %s media = %v, want exactly 10", b.Code, b.Average)
		}
	}
	if best[0].Code != "CODE-Recente" {
		t.Errorf("Melhores first = %s, want CODE-Recente", best[0].Code)
	}
}
