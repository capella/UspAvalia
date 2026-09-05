package middleware

import (
	"uspavalia/internal/models"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

// ScrapeMetricsCollector exposes the most recent Jupiter Web scrape runs on
// /metrics. The scraper writes a ScrapeRun row per execution; this collector
// reads the latest row per command on every scrape, so the numbers describe
// the last run rather than this process.
//
// The metric shapes are chosen for Grafana:
//   - every series carries a `command` label, so future scrapers (fetch-courses,
//     fetch-units) plug in without new metric names;
//   - status is a 0/1 gauge per `status` label value, which maps directly to a
//     Stat or State timeline panel with value mappings;
//   - timestamps are `*_timestamp_seconds` gauges, so `time() - metric` gives
//     the age of the last run and drives staleness alerts;
//   - page and discipline counts are single metrics split by a `result` or
//     `stage` label, so a stacked bar shows the whole funnel in one query;
//   - runs_total is a cumulative counter by status for `increase()` panels.
type ScrapeMetricsCollector struct {
	db *gorm.DB

	lastStatus            *prometheus.Desc
	lastStartedAt         *prometheus.Desc
	lastFinishedAt        *prometheus.Desc
	lastDuration          *prometheus.Desc
	lastSuccessFinishedAt *prometheus.Desc

	lastPagesRequested *prometheus.Desc
	lastPages          *prometheus.Desc
	lastDisciplines    *prometheus.Desc
	lastUnitsFound     *prometheus.Desc
	lastUnitListErrors *prometheus.Desc
	lastOfferings      *prometheus.Desc
	lastProfessors     *prometheus.Desc
	lastStoreErrors    *prometheus.Desc

	runsTotal *prometheus.Desc
}

// NewScrapeMetricsCollector creates a collector backed by the given database.
func NewScrapeMetricsCollector(db *gorm.DB) *ScrapeMetricsCollector {
	cmd := []string{"command"}
	return &ScrapeMetricsCollector{
		db: db,
		lastStatus: prometheus.NewDesc(
			"uspavalia_scrape_last_run_status",
			"Status of the most recent scrape run: 1 for the active status label, 0 for the others",
			[]string{"command", "status"}, nil,
		),
		lastStartedAt: prometheus.NewDesc(
			"uspavalia_scrape_last_run_started_timestamp_seconds",
			"Unix time at which the most recent scrape run started",
			cmd, nil,
		),
		lastFinishedAt: prometheus.NewDesc(
			"uspavalia_scrape_last_run_finished_timestamp_seconds",
			"Unix time at which the most recent scrape run finished (absent while running)",
			cmd, nil,
		),
		lastDuration: prometheus.NewDesc(
			"uspavalia_scrape_last_run_duration_seconds",
			"Wall-clock duration of the most recent finished scrape run",
			cmd, nil,
		),
		lastSuccessFinishedAt: prometheus.NewDesc(
			"uspavalia_scrape_last_success_timestamp_seconds",
			"Unix time at which the most recent scrape run that stored data (success or partial) finished",
			cmd, nil,
		),
		lastPagesRequested: prometheus.NewDesc(
			"uspavalia_scrape_last_run_pages_requested",
			"Jupiter Web pages requested by the most recent scrape run",
			cmd, nil,
		),
		lastPages: prometheus.NewDesc(
			"uspavalia_scrape_last_run_pages",
			"Jupiter Web pages by outcome in the most recent scrape run (loaded, http_error, bad_status, parse_error)",
			[]string{"command", "result"}, nil,
		),
		lastDisciplines: prometheus.NewDesc(
			"uspavalia_scrape_last_run_disciplines",
			"Disciplines by pipeline stage in the most recent scrape run (listed, processed, skipped, stored)",
			[]string{"command", "stage"}, nil,
		),
		lastUnitsFound: prometheus.NewDesc(
			"uspavalia_scrape_last_run_units_found",
			"Teaching units found by the most recent scrape run",
			cmd, nil,
		),
		lastUnitListErrors: prometheus.NewDesc(
			"uspavalia_scrape_last_run_unit_list_errors",
			"Teaching units whose discipline list could not be fetched in the most recent scrape run",
			cmd, nil,
		),
		lastOfferings: prometheus.NewDesc(
			"uspavalia_scrape_last_run_offerings_stored",
			"Class offerings created or updated by the most recent scrape run",
			cmd, nil,
		),
		lastProfessors: prometheus.NewDesc(
			"uspavalia_scrape_last_run_professors_created",
			"New professors created by the most recent scrape run",
			cmd, nil,
		),
		lastStoreErrors: prometheus.NewDesc(
			"uspavalia_scrape_last_run_store_errors",
			"Database write errors in the most recent scrape run",
			cmd, nil,
		),
		runsTotal: prometheus.NewDesc(
			"uspavalia_scrape_runs_total",
			"Total scrape runs recorded, by final status",
			[]string{"command", "status"}, nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *ScrapeMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.lastStatus
	ch <- c.lastStartedAt
	ch <- c.lastFinishedAt
	ch <- c.lastDuration
	ch <- c.lastSuccessFinishedAt
	ch <- c.lastPagesRequested
	ch <- c.lastPages
	ch <- c.lastDisciplines
	ch <- c.lastUnitsFound
	ch <- c.lastUnitListErrors
	ch <- c.lastOfferings
	ch <- c.lastProfessors
	ch <- c.lastStoreErrors
	ch <- c.runsTotal
}

// Collect implements prometheus.Collector
func (c *ScrapeMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	var commands []string
	if err := c.db.Model(&models.ScrapeRun{}).
		Distinct("command").
		Order("command").
		Pluck("command", &commands).Error; err != nil {
		return
	}

	for _, command := range commands {
		c.collectCommand(ch, command)
	}

	// Cumulative runs by status.
	var rows []struct {
		Command string
		Status  string
		Count   int64
	}
	if err := c.db.Model(&models.ScrapeRun{}).
		Select("command, status, COUNT(*) AS count").
		Group("command, status").
		Scan(&rows).Error; err == nil {
		for _, row := range rows {
			ch <- prometheus.MustNewConstMetric(
				c.runsTotal, prometheus.CounterValue, float64(row.Count), row.Command, row.Status,
			)
		}
	}
}

func (c *ScrapeMetricsCollector) collectCommand(ch chan<- prometheus.Metric, command string) {
	var last models.ScrapeRun
	if err := c.db.Where("command = ?", command).
		Order("started_at DESC, id DESC").
		First(&last).Error; err != nil {
		return
	}

	gauge := func(desc *prometheus.Desc, value float64, labels ...string) {
		ch <- prometheus.MustNewConstMetric(
			desc, prometheus.GaugeValue, value, append([]string{command}, labels...)...,
		)
	}

	for _, status := range models.ScrapeStatuses {
		v := 0.0
		if last.Status == status {
			v = 1
		}
		gauge(c.lastStatus, v, status)
	}

	gauge(c.lastStartedAt, float64(last.StartedAt.Unix()))
	if last.FinishedAt != nil {
		gauge(c.lastFinishedAt, float64(last.FinishedAt.Unix()))
		gauge(c.lastDuration, last.DurationSeconds)
	}

	gauge(c.lastPagesRequested, float64(last.PagesRequested))
	gauge(c.lastPages, float64(last.PagesLoaded), "loaded")
	gauge(c.lastPages, float64(last.HTTPErrors), "http_error")
	gauge(c.lastPages, float64(last.HTTPBadStatus), "bad_status")
	gauge(c.lastPages, float64(last.ParseErrors), "parse_error")

	gauge(c.lastDisciplines, float64(last.DisciplinesListed), "listed")
	gauge(c.lastDisciplines, float64(last.DisciplinesProcessed), "processed")
	gauge(c.lastDisciplines, float64(last.DisciplinesSkipped), "skipped")
	gauge(c.lastDisciplines, float64(last.DisciplinesStored), "stored")

	gauge(c.lastUnitsFound, float64(last.UnitsFound))
	gauge(c.lastUnitListErrors, float64(last.UnitListErrors))
	gauge(c.lastOfferings, float64(last.OfferingsStored))
	gauge(c.lastProfessors, float64(last.ProfessorsCreated))
	gauge(c.lastStoreErrors, float64(last.StoreErrors))

	// Latest run that actually refreshed data, for staleness alerting.
	var lastGood models.ScrapeRun
	if err := c.db.Where("command = ? AND status IN ? AND finished_at IS NOT NULL",
		command, []string{models.ScrapeStatusSuccess, models.ScrapeStatusPartial}).
		Order("finished_at DESC, id DESC").
		First(&lastGood).Error; err == nil && lastGood.FinishedAt != nil {
		gauge(c.lastSuccessFinishedAt, float64(lastGood.FinishedAt.Unix()))
	}
}
