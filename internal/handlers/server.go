package handlers

import (
	"html/template"
	"net/http"
	"time"
	"uspavalia/internal/config"
	"uspavalia/internal/middleware"
	"uspavalia/internal/models"
	"uspavalia/internal/services"

	csrf "filippo.io/csrf/gorilla"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

type Server struct {
	config       *config.Config
	db           *gorm.DB
	router       *mux.Router
	templates    *template.Template
	store        sessions.Store
	emailService *services.EmailService
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	store := sessions.NewCookieStore([]byte(cfg.Security.SecretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: false,
		Secure:   false, // Set to true in production with HTTPS
	}

	emailService := services.NewEmailService(cfg)

	// Register database metrics collector
	dbCollector := middleware.NewDatabaseMetricsCollector(db)
	prometheus.MustRegister(dbCollector)

	s := &Server{
		config:       cfg,
		db:           db,
		router:       mux.NewRouter(),
		store:        store,
		emailService: emailService,
	}

	s.loadTemplates()
	s.setupRoutes()
	s.startMetricsUpdater()
	s.startTokenCleanup()

	return s
}

func (s *Server) loadTemplates() {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	templates, err := template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		logrus.Printf("Warning: Failed to load templates: %v", err)
		s.templates = template.New("").Funcs(funcMap)
	} else {
		s.templates = templates
	}
}

func (s *Server) setupRoutes() {
	// Rate limiters
	generalLimiter := middleware.NewIPRateLimiter(rate.Every(time.Second), 30)
	authLimiter := middleware.NewIPRateLimiter(rate.Every(time.Minute), 30)
	apiLimiter := middleware.NewIPRateLimiter(rate.Every(time.Second*2), 20)

	// Start cleanup routines
	go generalLimiter.CleanupIPs()
	go authLimiter.CleanupIPs()
	go apiLimiter.CleanupIPs()

	disableCSRF := s.config.DevMode
	if disableCSRF {
		logrus.Warn("CSRF protection is DISABLED - do not use in production!")
	}

	CSRFMiddleware := csrf.Protect([]byte(s.config.Security.CSRFKey))
	s.router.Use(CSRFMiddleware)
	s.router.Use(middleware.SecurityHeaders)
	s.router.Use(middleware.Logging)
	s.router.Use(middleware.PrometheusMetrics)
	s.router.Use(middleware.RateLimit(generalLimiter))

	// Static files with caching headers
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))
	s.router.PathPrefix("/static/").Handler(middleware.StaticFileHeaders(staticHandler))

	// MatrUSP static files
	matruspHandler := http.StripPrefix("/matrusp/", http.FileServer(http.Dir("matrusp/")))
	s.router.PathPrefix("/matrusp/").Handler(middleware.StaticFileHeaders(matruspHandler))

	// Health check endpoint (no rate limiting)
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Prometheus metrics endpoint (no CSRF protection)
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Sitemap endpoint (no rate limiting, no CSRF protection)
	s.router.HandleFunc("/sitemap.xml", s.handleSitemap).Methods("GET")

	// Typeahead endpoint (no CSRF protection needed for search suggestions)
	s.router.HandleFunc("/typeahead", s.handleTypeahead).Methods("POST")

	// Vote activity API endpoint (for heatmap)
	s.router.HandleFunc("/api/vote-activity", s.handleVoteActivity).Methods("GET")

	// MatrUSP API endpoints
	s.router.HandleFunc("/api/matrusp/disciplines", s.handleMatruspDisciplines).Methods("GET")
	s.router.HandleFunc("/api/matrusp/discipline/{code}", s.handleMatruspDiscipline).Methods("GET")

	// MatrUSP redirect (handle /matrusp without trailing slash)
	s.router.HandleFunc("/matrusp", s.handleMatrusp).Methods("GET")

	// Public routes with optional authentication and CSRF protection
	public := s.router.PathPrefix("/").Subrouter()
	public.Use(middleware.OptionalAuth(s.store))

	// Authentication routes with strict rate limiting and CSRF protection
	auth := s.router.PathPrefix("/").Subrouter()
	auth.Use(middleware.StrictRateLimit(authLimiter))
	auth.Use(middleware.OptionalAuth(s.store))

	// Protected routes requiring authentication and CSRF protection
	protected := s.router.PathPrefix("/").Subrouter()
	protected.Use(middleware.RequireAuth(s.store))

	// Public pages
	// Handle legacy URLs with query parameters (/?p=page&id=123)
	// Must be before the simple "/" route
	public.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			s.handleLegacyURLs(w, r)
		} else {
			s.handleHome(w, r)
		}
	}).Methods("GET")
	public.HandleFunc("/disciplina/{id:[0-9]+}", s.handleDiscipline).Methods("GET")
	public.HandleFunc("/professor/{id:[0-9]+}", s.handleProfessor).Methods("GET")
	public.HandleFunc("/ver/{id:[0-9]+}", s.handleVer).Methods("GET")
	public.HandleFunc("/search", s.handleSearch).Methods("GET", "POST")
	public.HandleFunc("/sobre", s.handleAbout).Methods("GET")
	public.HandleFunc("/email", s.handleContact).Methods("GET", "POST")
	public.HandleFunc("/10melhores", s.handleTopRated).Methods("GET")
	public.HandleFunc("/destaques", s.handleTopRated).Methods("GET")

	// Authentication routes (rate limited)
	auth.HandleFunc("/login", s.handleRequestLogin).Methods("GET", "POST")
	auth.HandleFunc("/auth/magic-link", s.handleMagicLink).Methods("GET")
	auth.HandleFunc("/auth/google", s.handleGoogleLogin).Methods("GET")
	auth.HandleFunc("/auth/google/callback", s.handleGoogleCallback).Methods("GET")

	// Protected routes
	protected.HandleFunc("/logout", s.handleLogout).Methods("GET")

	// API routes with specific rate limiting and CSRF protection
	api := s.router.PathPrefix("/").Subrouter()
	api.Use(middleware.RateLimit(apiLimiter))
	api.Use(middleware.RequireAuth(s.store))

	// API endpoints (rate limited for actions)
	api.HandleFunc("/vote", s.handleVote).Methods("POST")
	api.HandleFunc("/vote-batch", s.handleBatchVote).Methods("POST")
	api.HandleFunc("/comment", s.handleComment).Methods("POST")
	api.HandleFunc("/vote-comment", s.handleCommentVote).Methods("POST")

	// Catch-all 404 handler - must be last
	s.router.NotFoundHandler = http.HandlerFunc(s.handle404)
}

func (s *Server) Router() http.Handler {
	return s.router
}

// startMetricsUpdater starts a background goroutine to update metrics periodically
func (s *Server) startMetricsUpdater() {
	// Update user count immediately
	s.updateUserCount()

	// Update user count every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			s.updateUserCount()
		}
	}()
}

// updateUserCount updates the current user count metric
func (s *Server) updateUserCount() {
	var count int64
	s.db.Model(&models.User{}).Count(&count)
	middleware.SetUsersCount(count)
}

// startTokenCleanup starts a background routine to clean up expired login tokens
func (s *Server) startTokenCleanup() {
	// Clean up expired tokens every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			result := s.db.Where("expires_at < ?", time.Now()).Delete(&models.LoginToken{})
			if result.Error != nil {
				logrus.Printf("Error cleaning up expired login tokens: %v", result.Error)
			} else if result.RowsAffected > 0 {
				logrus.Printf("Cleaned up %d expired login tokens", result.RowsAffected)
			}
		}
	}()
}

func (s *Server) Start() error {
	addr := s.config.Server.Host + ":" + s.config.Server.Port

	server := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.Server.IdleTimeout) * time.Second,
	}

	mode := "production"
	if s.config.DevMode {
		mode = "development"
	}
	logrus.Printf("Server starting on %s (mode: %s)", addr, mode)
	return server.ListenAndServe()
}

func (s *Server) renderTemplate(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	data interface{},
	extra_templates ...string,
) {
	// Parse templates fresh each time for proper inheritance
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Parse base template and the specific template
	files := []string{"templates/base.html", "templates/" + name + ".html"}
	files = append(files, extra_templates...)
	ts, err := template.New("").
		Funcs(funcMap).
		ParseFiles(files...)
	if err != nil {
		logrus.Printf("Error parsing templates: %v", err)
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Temos um problema :(")
		return
	}

	// Execute the base template
	if err := ts.ExecuteTemplate(w, "base", data); err != nil {
		logrus.Printf("Error executing template %s: %v", name, err)
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Temos um problema :(")
	}
}

func (s *Server) handle404(w http.ResponseWriter, r *http.Request) {
	s.renderErrorPage(w, r, http.StatusNotFound, "Página não encontrada")
}

func (s *Server) renderErrorPage(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	message string,
) {
	// Load and execute error template
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	ts, err := template.New("").
		Funcs(funcMap).
		ParseFiles("templates/base.html", "templates/error.html")
	if err != nil {
		logrus.Printf("Error parsing error templates: %v", err)
		// Fallback to basic error response
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(statusCode)
		w.Write(
			[]byte(`<html><body><h2>Temos um problema :(</h2><p>` + message + `</p></body></html>`),
		)
		return
	}

	data := struct {
		CSRFToken  string
		User       interface{}
		StatusCode int
		Message    string
		Data       map[string]interface{}
	}{
		StatusCode: statusCode,
		Message:    message,
		Data: map[string]interface{}{
			"StatusCode": statusCode,
			"Message":    message,
		},
	}

	// Set the HTTP status code before rendering the template
	w.WriteHeader(statusCode)

	if err := ts.ExecuteTemplate(w, "base", data); err != nil {
		logrus.Printf("Error executing error template: %v", err)
		// Final fallback
		w.Write(
			[]byte(`<html><body><h2>Temos um problema :(</h2><p>` + message + `</p></body></html>`),
		)
	}
}
