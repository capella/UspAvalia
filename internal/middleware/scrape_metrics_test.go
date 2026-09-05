package middleware

import (
	"strings"
	"testing"
	"time"
	"uspavalia/internal/models"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func scrapeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.ScrapeRun{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestScrapeMetricsEmptyDatabaseExposesNothing(t *testing.T) {
	c := NewScrapeMetricsCollector(scrapeTestDB(t))
	if n := testutil.CollectAndCount(c); n != 0 {
		t.Fatalf("expected no metrics without runs, got %d", n)
	}
}

func TestScrapeMetricsReportLatestRunAndLastSuccess(t *testing.T) {
	db := scrapeTestDB(t)

	t1 := time.Unix(1_700_000_000, 0)
	t1End := t1.Add(30 * time.Minute)
	t2 := t1.Add(7 * 24 * time.Hour)
	t2End := t2.Add(2 * time.Minute)

	runs := []models.ScrapeRun{
		{
			Command: "fetch-disciplines", Status: models.ScrapeStatusPartial,
			StartedAt: t1, FinishedAt: &t1End, DurationSeconds: 1800,
			PagesRequested: 100, PagesLoaded: 95, HTTPErrors: 3, HTTPBadStatus: 1, ParseErrors: 1,
			UnitsFound: 80, DisciplinesListed: 50, DisciplinesProcessed: 20, DisciplinesSkipped: 29,
			DisciplinesStored: 20, OfferingsStored: 40, ProfessorsCreated: 2,
		},
		{
			// Most recent run failed: last-success must still point at t1End.
			Command: "fetch-disciplines", Status: models.ScrapeStatusFailed,
			StartedAt: t2, FinishedAt: &t2End, DurationSeconds: 120,
			PagesRequested: 1, HTTPErrors: 1, ErrorMessage: "timeout",
		},
	}
	if err := db.Create(&runs).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	expected := `
# HELP uspavalia_scrape_last_run_status Status of the most recent scrape run: 1 for the active status label, 0 for the others
# TYPE uspavalia_scrape_last_run_status gauge
uspavalia_scrape_last_run_status{command="fetch-disciplines",status="failed"} 1
uspavalia_scrape_last_run_status{command="fetch-disciplines",status="partial"} 0
uspavalia_scrape_last_run_status{command="fetch-disciplines",status="running"} 0
uspavalia_scrape_last_run_status{command="fetch-disciplines",status="success"} 0
# HELP uspavalia_scrape_last_run_started_timestamp_seconds Unix time at which the most recent scrape run started
# TYPE uspavalia_scrape_last_run_started_timestamp_seconds gauge
uspavalia_scrape_last_run_started_timestamp_seconds{command="fetch-disciplines"} 1.7006048e+09
# HELP uspavalia_scrape_last_run_finished_timestamp_seconds Unix time at which the most recent scrape run finished (absent while running)
# TYPE uspavalia_scrape_last_run_finished_timestamp_seconds gauge
uspavalia_scrape_last_run_finished_timestamp_seconds{command="fetch-disciplines"} 1.70060492e+09
# HELP uspavalia_scrape_last_run_duration_seconds Wall-clock duration of the most recent finished scrape run
# TYPE uspavalia_scrape_last_run_duration_seconds gauge
uspavalia_scrape_last_run_duration_seconds{command="fetch-disciplines"} 120
# HELP uspavalia_scrape_last_success_timestamp_seconds Unix time at which the most recent scrape run that stored data (success or partial) finished
# TYPE uspavalia_scrape_last_success_timestamp_seconds gauge
uspavalia_scrape_last_success_timestamp_seconds{command="fetch-disciplines"} 1.7000018e+09
# HELP uspavalia_scrape_last_run_pages Jupiter Web pages by outcome in the most recent scrape run (loaded, http_error, bad_status, parse_error)
# TYPE uspavalia_scrape_last_run_pages gauge
uspavalia_scrape_last_run_pages{command="fetch-disciplines",result="bad_status"} 0
uspavalia_scrape_last_run_pages{command="fetch-disciplines",result="http_error"} 1
uspavalia_scrape_last_run_pages{command="fetch-disciplines",result="loaded"} 0
uspavalia_scrape_last_run_pages{command="fetch-disciplines",result="parse_error"} 0
# HELP uspavalia_scrape_runs_total Total scrape runs recorded, by final status
# TYPE uspavalia_scrape_runs_total counter
uspavalia_scrape_runs_total{command="fetch-disciplines",status="failed"} 1
uspavalia_scrape_runs_total{command="fetch-disciplines",status="partial"} 1
`
	c := NewScrapeMetricsCollector(db)
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"uspavalia_scrape_last_run_status",
		"uspavalia_scrape_last_run_started_timestamp_seconds",
		"uspavalia_scrape_last_run_finished_timestamp_seconds",
		"uspavalia_scrape_last_run_duration_seconds",
		"uspavalia_scrape_last_success_timestamp_seconds",
		"uspavalia_scrape_last_run_pages",
		"uspavalia_scrape_runs_total",
	); err != nil {
		t.Fatal(err)
	}
}

func TestScrapeMetricsRunningRunHasNoFinishedTimestamp(t *testing.T) {
	db := scrapeTestDB(t)
	if err := db.Create(&models.ScrapeRun{
		Command: "fetch-disciplines", Status: models.ScrapeStatusRunning, StartedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	c := NewScrapeMetricsCollector(db)
	if n := testutil.CollectAndCount(c, "uspavalia_scrape_last_run_finished_timestamp_seconds",
		"uspavalia_scrape_last_run_duration_seconds",
		"uspavalia_scrape_last_success_timestamp_seconds"); n != 0 {
		t.Fatalf("running run should expose no finish/duration/success series, got %d", n)
	}
	if n := testutil.CollectAndCount(c, "uspavalia_scrape_last_run_status"); n != len(models.ScrapeStatuses) {
		t.Fatalf("expected one status series per status, got %d", n)
	}
}
