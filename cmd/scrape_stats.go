package cmd

import (
	"fmt"
	"sync/atomic"
	"time"
	"uspavalia/internal/models"

	"gorm.io/gorm"
)

// scrapeStats holds process-wide counters for one scraper run. The HTTP
// helpers in utils.go feed the page counters, and the fetch commands feed the
// pipeline counters. Everything is atomic because discipline pages are
// fetched from many goroutines at once.
type scrapeStats struct {
	pagesRequested atomic.Int64
	pagesLoaded    atomic.Int64
	httpErrors     atomic.Int64
	httpBadStatus  atomic.Int64
	parseErrors    atomic.Int64

	unitsFound           atomic.Int64
	unitListErrors       atomic.Int64
	disciplinesListed    atomic.Int64
	disciplinesProcessed atomic.Int64
	disciplinesSkipped   atomic.Int64
	disciplinesStored    atomic.Int64
	offeringsStored      atomic.Int64
	professorsCreated    atomic.Int64
	storeErrors          atomic.Int64
}

// stats is the single collector for the current process. A CLI invocation
// runs exactly one scraper command, so a package-level instance is enough.
var stats = &scrapeStats{}

// recordHTTPResult classifies the outcome of one page fetch.
func (s *scrapeStats) recordHTTPResult(statusCode int, err error) {
	s.pagesRequested.Add(1)
	switch {
	case err != nil && statusCode == 0:
		s.httpErrors.Add(1)
	case err != nil:
		// Body arrived but could not be decoded or parsed as HTML.
		s.parseErrors.Add(1)
	case statusCode != 200:
		s.httpBadStatus.Add(1)
	default:
		s.pagesLoaded.Add(1)
	}
}

// applyTo copies the counters into a ScrapeRun row.
func (s *scrapeStats) applyTo(run *models.ScrapeRun) {
	run.PagesRequested = int(s.pagesRequested.Load())
	run.PagesLoaded = int(s.pagesLoaded.Load())
	run.HTTPErrors = int(s.httpErrors.Load())
	run.HTTPBadStatus = int(s.httpBadStatus.Load())
	run.ParseErrors = int(s.parseErrors.Load())
	run.UnitsFound = int(s.unitsFound.Load())
	run.UnitListErrors = int(s.unitListErrors.Load())
	run.DisciplinesListed = int(s.disciplinesListed.Load())
	run.DisciplinesProcessed = int(s.disciplinesProcessed.Load())
	run.DisciplinesSkipped = int(s.disciplinesSkipped.Load())
	run.DisciplinesStored = int(s.disciplinesStored.Load())
	run.OfferingsStored = int(s.offeringsStored.Load())
	run.ProfessorsCreated = int(s.professorsCreated.Load())
	run.StoreErrors = int(s.storeErrors.Load())
}

// finalStatus derives the run outcome from the counters. A fatal error is
// always "failed"; a run that stored nothing is "failed" too, since the site
// would otherwise keep serving stale data with a green-looking record; any
// per-page or per-row error makes the run "partial".
func (s *scrapeStats) finalStatus(fatal error) string {
	if fatal != nil {
		return models.ScrapeStatusFailed
	}
	if s.disciplinesStored.Load() == 0 {
		return models.ScrapeStatusFailed
	}
	if s.httpErrors.Load() > 0 || s.httpBadStatus.Load() > 0 || s.parseErrors.Load() > 0 ||
		s.unitListErrors.Load() > 0 || s.storeErrors.Load() > 0 {
		return models.ScrapeStatusPartial
	}
	return models.ScrapeStatusSuccess
}

// scrapeRecorder persists the ScrapeRun row for the current command. A nil
// recorder is valid and does nothing, so callers can use it unconditionally.
type scrapeRecorder struct {
	db  *gorm.DB
	run models.ScrapeRun
}

// startScrapeRun inserts a "running" row so a crash mid-run is visible as a
// run that never finished.
func startScrapeRun(db *gorm.DB, command string) (*scrapeRecorder, error) {
	r := &scrapeRecorder{
		db: db,
		run: models.ScrapeRun{
			Command:   command,
			Status:    models.ScrapeStatusRunning,
			StartedAt: time.Now(),
		},
	}
	if err := db.Create(&r.run).Error; err != nil {
		return nil, fmt.Errorf("record scrape start: %w", err)
	}
	return r, nil
}

// finish writes the final counters and status. It prints instead of
// returning an error because it runs on the way out and there is nothing
// left to do about a failure.
func (r *scrapeRecorder) finish(s *scrapeStats, fatal error) {
	if r == nil {
		return
	}
	now := time.Now()
	r.run.FinishedAt = &now
	r.run.DurationSeconds = now.Sub(r.run.StartedAt).Seconds()
	r.run.Status = s.finalStatus(fatal)
	if fatal != nil {
		r.run.ErrorMessage = fatal.Error()
	} else if r.run.Status == models.ScrapeStatusFailed {
		r.run.ErrorMessage = "no disciplines were stored"
	}
	s.applyTo(&r.run)

	// Select(*) forces zero values (e.g. 0 errors) to be written too.
	if err := r.db.Model(&r.run).Select("*").Updates(&r.run).Error; err != nil {
		fmt.Printf("Warning: Failed to record scrape run: %v\n", err)
		return
	}
	fmt.Printf("- Scrape run #%d recorded with status %q -\n", r.run.ID, r.run.Status)
}
