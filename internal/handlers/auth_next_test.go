package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
	"uspavalia/internal/config"
	"uspavalia/internal/database"
	"uspavalia/internal/middleware"
	"uspavalia/internal/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSafeNextPath(t *testing.T) {
	cases := map[string]string{
		"":                            "",
		"/":                           "/",
		"/disciplina/4953#modal27661": "/disciplina/4953#modal27661",
		"/ver/1?x=1#modal1":           "/ver/1?x=1#modal1",
		"//evil.com/x":                "",
		"/\\evil.com":                 "",
		"http://evil.com/":            "",
		"https://uspavalia.com/x":     "",
		"javascript:alert(1)":         "",
		"disciplina/1":                "",
		"/a\r\nSet-Cookie: x=y":       "",
	}
	for in, want := range cases {
		if got := safeNextPath(in); got != want {
			t.Errorf("safeNextPath(%q) = %q, want %q", in, got, want)
		}
	}
}

// newAuthServer builds a Server able to render pages and log users in.
func newAuthServer(t *testing.T) *Server {
	t.Helper()
	// Templates are resolved relative to the repository root.
	wd, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Unit{}, &models.Discipline{}, &models.Professor{},
		&models.ClassProfessor{}, &models.Vote{}, &models.User{}, &models.LoginToken{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.CreateViews(db); err != nil {
		t.Fatalf("create views: %v", err)
	}
	cfg := &config.Config{}
	cfg.Security.SessionName = "uspavalia_session"
	cfg.Security.SecretKey = "test-secret"
	return &Server{db: db, config: cfg, store: sessions.NewCookieStore([]byte("test-secret"))}
}

func seedClassProfessor(t *testing.T, db *gorm.DB) models.ClassProfessor {
	t.Helper()
	unit := models.Unit{Name: "IME"}
	db.Create(&unit)
	disc := models.Discipline{Name: "Introdução à Computação", Code: "MAC0110", UnitID: unit.ID}
	db.Create(&disc)
	prof := models.Professor{Name: "Leliane Nunes de Barros", UnitID: unit.ID}
	db.Create(&prof)
	cp := models.ClassProfessor{ClassID: disc.ID, ProfessorID: prof.ID}
	db.Create(&cp)
	return cp
}

func TestMagicLinkRedirectsToRememberedNext(t *testing.T) {
	s := newAuthServer(t)

	// The login page remembers where the user wants to go.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login?next=%2Fdisciplina%2F4953%23modal27661", nil)
	s.handleRequestLogin(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /login status = %d, want 200", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login page did not set a session cookie")
	}

	// The magic link, opened in the same browser, sends the user back there.
	s.db.Create(&models.LoginToken{Token: "tok", EmailHash: "h", ExpiresAt: time.Now().Add(time.Hour)})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/auth/magic-link?token=tok", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	s.handleMagicLink(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("magic link status = %d, want 303", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/disciplina/4953#modal27661" {
		t.Errorf("Location = %q, want the discipline page with the modal hash", got)
	}
}

func TestMagicLinkWithoutNextGoesHome(t *testing.T) {
	s := newAuthServer(t)
	s.db.Create(&models.LoginToken{Token: "tok", EmailHash: "h", ExpiresAt: time.Now().Add(time.Hour)})
	rec := httptest.NewRecorder()
	s.handleMagicLink(rec, httptest.NewRequest(http.MethodGet, "/auth/magic-link?token=tok", nil))
	if got := rec.Header().Get("Location"); got != "/" {
		t.Errorf("Location = %q, want /", got)
	}
}

func disciplineRequest(loggedIn bool) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/disciplina/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	if loggedIn {
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "1"))
	}
	return req
}

func TestDisciplinePageModalWiring(t *testing.T) {
	s := newAuthServer(t)
	seedClassProfessor(t, s.db)
	s.db.Create(&models.User{EmailHash: "u"})

	// Logged out: the button is a login link that comes back to this modal.
	rec := httptest.NewRecorder()
	s.handleDiscipline(rec, disciplineRequest(false))
	body := rec.Body.String()
	wantLink := `href="/login?next=%2fdisciplina%2f1%23modal1"`
	if !strings.Contains(strings.ToLower(body), wantLink) {
		t.Errorf("logged-out page lacks %s", wantLink)
	}
	if strings.Contains(body, `id="modal1"`) {
		t.Errorf("logged-out page should not render the rating modal")
	}
	// Logged-out visitors see the same green "Avaliar" button, not a
	// special login label.
	if strings.Contains(body, "Login para avaliar") || !strings.Contains(body, `class="btn btn-success"`) {
		t.Errorf("logged-out page should show the green Avaliar button")
	}

	// Logged in: the button targets the modal for this class-professor,
	// and the modal names the discipline.
	rec = httptest.NewRecorder()
	s.handleDiscipline(rec, disciplineRequest(true))
	body = rec.Body.String()
	for _, want := range []string{
		`data-target="#modal1"`,
		`id="modal1"`,
		"Leliane Nunes de Barros -\n            MAC0110",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("logged-in page lacks %q", want)
		}
	}
	if strings.Contains(body, `data-target="#modal"`) {
		t.Errorf("found a button whose modal target has no id")
	}
	// A <form> inside <tbody> is emptied by the HTML parser, so the modal
	// must come after the table.
	if strings.Index(body, `id="modal1"`) < strings.Index(body, "</table>") {
		t.Errorf("rating modal is rendered inside the table")
	}
}

func TestRequireAuthRedirectsWithNextAndRejectsAPI(t *testing.T) {
	store := sessions.NewCookieStore([]byte("k"))
	h := middleware.RequireAuth(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/logout?a=1", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/login?next=%2Flogout%3Fa%3D1" {
		t.Errorf("browser request: code=%d location=%q", rec.Code, rec.Header().Get("Location"))
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/vote-batch", nil)
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("json request: code=%d, want 401", rec.Code)
	}
}
