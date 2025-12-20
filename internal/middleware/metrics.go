package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uspavalia_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uspavalia_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uspavalia_http_errors_total",
			Help: "Total number of HTTP errors (4xx and 5xx)",
		},
		[]string{"method", "path", "status"},
	)

	// Application metrics
	votesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "uspavalia_votes_total",
			Help: "Total number of votes submitted",
		},
	)

	commentsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "uspavalia_comments_total",
			Help: "Total number of comments submitted",
		},
	)

	registrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "uspavalia_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	usersCount = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "uspavalia_users_count",
			Help: "Current number of registered users",
		},
		func() float64 {
			// This will be updated via SetUsersCount
			return float64(currentUsersCount)
		},
	)
)

var currentUsersCount int64

// metricsResponseWriter wraps http.ResponseWriter to capture status code for metrics
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// PrometheusMetrics middleware records HTTP metrics
func PrometheusMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		rw := newMetricsResponseWriter(w)

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get route pattern
		route := r.URL.Path
		if routeMatch := mux.CurrentRoute(r); routeMatch != nil {
			if pathTemplate, err := routeMatch.GetPathTemplate(); err == nil {
				route = pathTemplate
			}
		}

		statusCode := strconv.Itoa(rw.statusCode)
		method := r.Method

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, route, statusCode).Inc()
		httpRequestDuration.WithLabelValues(method, route, statusCode).Observe(duration)

		// Record errors (4xx and 5xx)
		if rw.statusCode >= 400 {
			httpErrorsTotal.WithLabelValues(method, route, statusCode).Inc()
		}
	})
}

// RecordVote increments the vote counter
func RecordVote() {
	votesTotal.Inc()
}

// RecordComment increments the comment counter
func RecordComment() {
	commentsTotal.Inc()
}

// RecordRegistration increments the registration counter
func RecordRegistration() {
	registrationsTotal.Inc()
}

// SetUsersCount sets the current number of users
func SetUsersCount(count int64) {
	currentUsersCount = count
}
