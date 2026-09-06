package models

import "time"

// Scrape run status values.
const (
	ScrapeStatusRunning = "running"
	ScrapeStatusSuccess = "success"
	ScrapeStatusPartial = "partial"
	ScrapeStatusFailed  = "failed"
)

// ScrapeStatuses lists every status a ScrapeRun can have, in a stable order.
var ScrapeStatuses = []string{
	ScrapeStatusRunning,
	ScrapeStatusSuccess,
	ScrapeStatusPartial,
	ScrapeStatusFailed,
}

// ScrapeRun records one execution of a Jupiter Web scraper command. The
// scraper runs as a separate process from the web server, so this row is the
// channel through which the server exposes scrape health on /metrics.
type ScrapeRun struct {
	ID      uint   `gorm:"primaryKey"             json:"id"`
	Command string `gorm:"size:50;not null;index" json:"command"`
	Status  string `gorm:"size:20;not null"       json:"status"`

	StartedAt       time.Time  `gorm:"not null" json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	DurationSeconds float64    `json:"duration_seconds"`

	// HTTP activity against Jupiter Web.
	PagesRequested int `json:"pages_requested"` // GET requests attempted
	PagesLoaded    int `json:"pages_loaded"`    // HTTP 200 with a parseable body
	HTTPErrors     int `json:"http_errors"`     // transport errors and timeouts
	HTTPBadStatus  int `json:"http_bad_status"` // non-200 responses
	ParseErrors    int `json:"parse_errors"`    // pages that loaded but yielded no usable data

	// Pipeline counters.
	UnitsFound           int `json:"units_found"`
	UnitListErrors       int `json:"unit_list_errors"` // unit pages whose discipline list failed
	DisciplinesListed    int `json:"disciplines_listed"`
	DisciplinesProcessed int `json:"disciplines_processed"`
	DisciplinesSkipped   int `json:"disciplines_skipped"` // no classes offered; not an error
	DisciplinesStored    int `json:"disciplines_stored"`
	OfferingsStored      int `json:"offerings_stored"`
	ProfessorsCreated    int `json:"professors_created"`
	StoreErrors          int `json:"store_errors"`

	ErrorMessage string `gorm:"type:text" json:"error_message"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
