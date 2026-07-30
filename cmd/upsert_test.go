package cmd

import (
	"testing"
	"uspavalia/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Unit{},
		&models.Professor{},
		&models.Discipline{},
		&models.ClassProfessor{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func countRows(t *testing.T, db *gorm.DB, model interface{}) int64 {
	t.Helper()
	var n int64
	if err := db.Model(model).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// Whitespace variants of the same unit name must resolve to a single row.
// The production units table once held 6,430 rows for 80 real units.
func TestGetOrCreateUnitDedupesWhitespaceVariants(t *testing.T) {
	db := testDB(t)

	for _, name := range []string{"Escola Politécnica", " Escola Politécnica", "Escola Politécnica ", "  Escola Politécnica  "} {
		unit, err := getOrCreateUnit(db, name)
		if err != nil {
			t.Fatalf("getOrCreateUnit(%q): %v", name, err)
		}
		if unit.Name != "Escola Politécnica" {
			t.Errorf("getOrCreateUnit(%q) name = %q, want trimmed", name, unit.Name)
		}
	}
	if n := countRows(t, db, &models.Unit{}); n != 1 {
		t.Errorf("units rows = %d, want 1", n)
	}
}

func TestGetOrCreateUnitRejectsEmptyName(t *testing.T) {
	db := testDB(t)
	if _, err := getOrCreateUnit(db, "   "); err == nil {
		t.Error("expected error for blank unit name")
	}
	if n := countRows(t, db, &models.Unit{}); n != 0 {
		t.Errorf("units rows = %d, want 0", n)
	}
}

func TestGetOrCreateProfessorDedupesWhitespaceVariants(t *testing.T) {
	db := testDB(t)

	first, created, err := getOrCreateProfessor(db, "Alfredo Goldman Vel Lejbman", 1)
	if err != nil || !created {
		t.Fatalf("first create: created=%v err=%v", created, err)
	}
	second, created, err := getOrCreateProfessor(db, " Alfredo Goldman Vel Lejbman ", 1)
	if err != nil {
		t.Fatalf("second lookup: %v", err)
	}
	if created {
		t.Error("whitespace variant created a second professor")
	}
	if second.ID != first.ID {
		t.Errorf("variant resolved to id %d, want %d", second.ID, first.ID)
	}
	if n := countRows(t, db, &models.Professor{}); n != 1 {
		t.Errorf("professors rows = %d, want 1", n)
	}
}

// A discipline is identified by code: a renamed discipline must update the
// existing row, never create a second one with the same code.
func TestGetOrCreateDisciplineKeyedByCode(t *testing.T) {
	db := testDB(t)

	first, err := getOrCreateDiscipline(db, models.Discipline{Code: "4323102", Name: "Física II", UnitID: 1})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	renamed, err := getOrCreateDiscipline(db, models.Discipline{Code: "4323102", Name: "Física 2", UnitID: 1})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if renamed.ID != first.ID {
		t.Errorf("rename resolved to id %d, want %d", renamed.ID, first.ID)
	}
	if n := countRows(t, db, &models.Discipline{}); n != 1 {
		t.Errorf("disciplines rows = %d, want 1", n)
	}

	var stored models.Discipline
	if err := db.First(&stored, first.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if stored.Name != "Física 2" {
		t.Errorf("stored name = %q, want updated name", stored.Name)
	}
}

func TestGetOrCreateClassProfessorIsIdempotent(t *testing.T) {
	db := testDB(t)

	_, created, err := getOrCreateClassProfessor(db, 10, 20)
	if err != nil || !created {
		t.Fatalf("first create: created=%v err=%v", created, err)
	}
	_, created, err = getOrCreateClassProfessor(db, 10, 20)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if created {
		t.Error("second call created a duplicate association")
	}
	if n := countRows(t, db, &models.ClassProfessor{}); n != 1 {
		t.Errorf("class_professors rows = %d, want 1", n)
	}
}

// A query failure (e.g. a wrong column name) must surface as an error, not be
// treated as "not found": fetch-courses once queried the legacy NOME column
// and created a new unit per course per run.
func TestGetOrCreateUnitPropagatesQueryErrors(t *testing.T) {
	db := testDB(t)

	// Point the model at a table whose schema doesn't match to force a
	// column error on lookup.
	if err := db.Exec("DROP TABLE units").Error; err != nil {
		t.Fatalf("drop: %v", err)
	}
	if err := db.Exec("CREATE TABLE units (id integer primary key, nome text)").Error; err != nil {
		t.Fatalf("recreate: %v", err)
	}

	if _, err := getOrCreateUnit(db, "Escola Politécnica"); err == nil {
		t.Fatal("expected query error to propagate, got nil (would have created a duplicate)")
	}
	if n := countRows(t, db, &models.Unit{}); n != 0 {
		t.Errorf("units rows = %d, want 0 (nothing should be created on error)", n)
	}
}

// The schema itself must refuse duplicate identities once migrated.
func TestUniqueIndexesEnforced(t *testing.T) {
	db := testDB(t)

	if err := db.Create(&models.Professor{Name: "Ada Lovelace", UnitID: 1}).Error; err != nil {
		t.Fatalf("professor: %v", err)
	}
	if err := db.Create(&models.Professor{Name: "Ada Lovelace", UnitID: 1}).Error; err == nil {
		t.Error("duplicate professor name accepted; unique index missing")
	}

	if err := db.Create(&models.Discipline{Code: "MAC0219", Name: "PCP", UnitID: 1}).Error; err != nil {
		t.Fatalf("discipline: %v", err)
	}
	if err := db.Create(&models.Discipline{Code: "MAC0219", Name: "Outra", UnitID: 1}).Error; err == nil {
		t.Error("duplicate discipline code accepted; unique index missing")
	}

	if err := db.Create(&models.ClassProfessor{ClassID: 1, ProfessorID: 2}).Error; err != nil {
		t.Fatalf("class_professor: %v", err)
	}
	if err := db.Create(&models.ClassProfessor{ClassID: 1, ProfessorID: 2}).Error; err == nil {
		t.Error("duplicate (class_id, professor_id) accepted; unique index missing")
	}
}
