package cmd

import (
	"errors"
	"testing"
	"uspavalia/internal/models"
)

func TestScrapeStatsFinalStatus(t *testing.T) {
	cases := []struct {
		name  string
		setup func(s *scrapeStats)
		fatal error
		want  string
	}{
		{"fatal error", func(s *scrapeStats) { s.disciplinesStored.Store(10) }, errors.New("boom"), models.ScrapeStatusFailed},
		{"nothing stored", func(s *scrapeStats) { s.pagesLoaded.Store(5) }, nil, models.ScrapeStatusFailed},
		{"clean run", func(s *scrapeStats) { s.disciplinesStored.Store(10) }, nil, models.ScrapeStatusSuccess},
		{"http errors", func(s *scrapeStats) { s.disciplinesStored.Store(10); s.httpErrors.Store(1) }, nil, models.ScrapeStatusPartial},
		{"parse errors", func(s *scrapeStats) { s.disciplinesStored.Store(10); s.parseErrors.Store(1) }, nil, models.ScrapeStatusPartial},
		{"unit list errors", func(s *scrapeStats) { s.disciplinesStored.Store(10); s.unitListErrors.Store(1) }, nil, models.ScrapeStatusPartial},
		{"store errors", func(s *scrapeStats) { s.disciplinesStored.Store(10); s.storeErrors.Store(1) }, nil, models.ScrapeStatusPartial},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &scrapeStats{}
			tc.setup(s)
			if got := s.finalStatus(tc.fatal); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestScrapeStatsRecordHTTPResult(t *testing.T) {
	s := &scrapeStats{}
	s.recordHTTPResult(200, nil)
	s.recordHTTPResult(200, nil)
	s.recordHTTPResult(0, errors.New("dial timeout"))
	s.recordHTTPResult(500, nil)
	s.recordHTTPResult(200, errors.New("bad charset"))

	var run models.ScrapeRun
	s.applyTo(&run)
	if run.PagesRequested != 5 || run.PagesLoaded != 2 || run.HTTPErrors != 1 ||
		run.HTTPBadStatus != 1 || run.ParseErrors != 1 {
		t.Fatalf("unexpected counters: %+v", run)
	}
}

// A run is inserted as "running" at start and rewritten with counters and a
// final status at the end, so a crash leaves a visible unfinished row.
func TestScrapeRecorderLifecycle(t *testing.T) {
	db := testDB(t)
	if err := db.AutoMigrate(&models.ScrapeRun{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	rec, err := startScrapeRun(db, "fetch-disciplines")
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	var started models.ScrapeRun
	if err := db.First(&started, rec.run.ID).Error; err != nil {
		t.Fatalf("load: %v", err)
	}
	if started.Status != models.ScrapeStatusRunning || started.FinishedAt != nil {
		t.Fatalf("expected a running row, got %+v", started)
	}

	s := &scrapeStats{}
	s.recordHTTPResult(200, nil)
	s.disciplinesStored.Store(3)
	s.httpBadStatus.Store(2)
	rec.finish(s, nil)

	var finished models.ScrapeRun
	if err := db.First(&finished, rec.run.ID).Error; err != nil {
		t.Fatalf("load: %v", err)
	}
	if finished.Status != models.ScrapeStatusPartial {
		t.Fatalf("status = %q, want partial", finished.Status)
	}
	if finished.FinishedAt == nil || finished.DisciplinesStored != 3 ||
		finished.PagesRequested != 1 || finished.HTTPBadStatus != 2 {
		t.Fatalf("counters not persisted: %+v", finished)
	}
	if n := countRows(t, db, &models.ScrapeRun{}); n != 1 {
		t.Fatalf("expected a single row, got %d", n)
	}

	// Fatal path records the message.
	rec2, _ := startScrapeRun(db, "fetch-disciplines")
	rec2.finish(&scrapeStats{}, errors.New("jupiter unreachable"))
	var failed models.ScrapeRun
	db.First(&failed, rec2.run.ID)
	if failed.Status != models.ScrapeStatusFailed || failed.ErrorMessage != "jupiter unreachable" {
		t.Fatalf("fatal run not recorded: %+v", failed)
	}

	// A nil recorder (preview mode) must be a no-op.
	var none *scrapeRecorder
	none.finish(&scrapeStats{}, nil)
}
