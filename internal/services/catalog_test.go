package services

import (
	"testing"
	"time"
	"uspavalia/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCurrentSemester(t *testing.T) {
	cases := map[string]Semester{
		"2026-01-01": {2026, 1},
		"2026-06-30": {2026, 1},
		"2026-07-01": {2026, 2},
		"2026-12-31": {2026, 2},
	}
	for date, want := range cases {
		d, _ := time.Parse("2006-01-02", date)
		if got := CurrentSemester(d); got != want {
			t.Errorf("CurrentSemester(%s) = %v, want %v", date, got, want)
		}
	}
}

func TestSemesterCodeRoundTrip(t *testing.T) {
	sem := Semester{2026, 2}
	if sem.Code() != "2026200" || sem.Prefix() != "20262" || sem.String() != "2026/2" {
		t.Fatalf("code=%s prefix=%s string=%s", sem.Code(), sem.Prefix(), sem.String())
	}
	if got, ok := ParseSemesterCode("2026241"); !ok || got != sem {
		t.Errorf("ParseSemesterCode(2026241) = %v %v", got, ok)
	}
	for _, bad := range []string{"", "2026", "2026341", "abcd141"} {
		if _, ok := ParseSemesterCode(bad); ok {
			t.Errorf("ParseSemesterCode(%q) accepted", bad)
		}
	}
}

func TestSemesterValid(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2026-09-05")
	if !(Semester{2026, 1}).Valid(now) || !(Semester{2027, 2}).Valid(now) {
		t.Error("valid semesters rejected")
	}
	for _, bad := range []Semester{{1999, 1}, {2028, 1}, {2026, 0}, {2026, 3}} {
		if bad.Valid(now) {
			t.Errorf("%v accepted", bad)
		}
	}
}

func TestProfessorSemestersDedupesSlotsAndTurmas(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.ClassOffering{}); err != nil {
		t.Fatal(err)
	}
	rows := []models.ClassOffering{
		// Same professor in two time slots of one turma.
		{DisciplineID: 1, Code: "2026241", Schedules: `[{"dia":"seg","professores":["Ana"]},{"dia":"qua","professores":["Ana"]}]`},
		// Same semester, another turma.
		{DisciplineID: 1, Code: "2026242", Schedules: `[{"dia":"ter","professores":["Ana","Bia"]}]`},
		// Older semester, plus one where Ana does not teach.
		{DisciplineID: 1, Code: "2025141", Schedules: `[{"dia":"ter","professores":["Ana"]}]`},
		{DisciplineID: 1, Code: "2025241", Schedules: `[{"dia":"ter","professores":["Bia"]}]`},
		// Another discipline must not leak in.
		{DisciplineID: 2, Code: "2024141", Schedules: `[{"dia":"ter","professores":["Ana"]}]`},
	}
	for i := range rows {
		if err := db.Create(&rows[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	got := ProfessorSemesters(db, 1, "Ana")
	want := []Semester{{2026, 2}, {2025, 1}}
	if len(got) != len(want) {
		t.Fatalf("ProfessorSemesters = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ProfessorSemesters = %v, want %v", got, want)
		}
	}
}
