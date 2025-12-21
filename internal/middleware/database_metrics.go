package middleware

import (
	"uspavalia/internal/models"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

// DatabaseMetricsCollector collects database-related metrics
type DatabaseMetricsCollector struct {
	db *gorm.DB

	totalVotes       *prometheus.Desc
	totalUsers       *prometheus.Desc
	totalComments    *prometheus.Desc
	totalDisciplines *prometheus.Desc
	totalProfessors  *prometheus.Desc
}

// NewDatabaseMetricsCollector creates a new database metrics collector
func NewDatabaseMetricsCollector(db *gorm.DB) *DatabaseMetricsCollector {
	return &DatabaseMetricsCollector{
		db: db,
		totalVotes: prometheus.NewDesc(
			"uspavalia_total_votes",
			"Total number of votes in database",
			nil, nil,
		),
		totalUsers: prometheus.NewDesc(
			"uspavalia_total_users",
			"Total number of registered users",
			nil, nil,
		),
		totalComments: prometheus.NewDesc(
			"uspavalia_total_comments",
			"Total number of comments",
			nil, nil,
		),
		totalDisciplines: prometheus.NewDesc(
			"uspavalia_total_disciplines",
			"Total number of disciplines",
			nil, nil,
		),
		totalProfessors: prometheus.NewDesc(
			"uspavalia_total_professors",
			"Total number of professors",
			nil, nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *DatabaseMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalVotes
	ch <- c.totalUsers
	ch <- c.totalComments
	ch <- c.totalDisciplines
	ch <- c.totalProfessors
}

// Collect implements prometheus.Collector
func (c *DatabaseMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	var count int64

	// Count votes
	if err := c.db.Model(&models.Vote{}).Count(&count).Error; err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.totalVotes,
			prometheus.GaugeValue,
			float64(count),
		)
	}

	// Count users
	if err := c.db.Model(&models.User{}).Count(&count).Error; err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.totalUsers,
			prometheus.GaugeValue,
			float64(count),
		)
	}

	// Count comments
	if err := c.db.Model(&models.Comment{}).Count(&count).Error; err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.totalComments,
			prometheus.GaugeValue,
			float64(count),
		)
	}

	// Count disciplines
	if err := c.db.Model(&models.Discipline{}).Count(&count).Error; err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.totalDisciplines,
			prometheus.GaugeValue,
			float64(count),
		)
	}

	// Count professors
	if err := c.db.Model(&models.Professor{}).Count(&count).Error; err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.totalProfessors,
			prometheus.GaugeValue,
			float64(count),
		)
	}
}
