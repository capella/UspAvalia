package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"uspavalia/internal/config"
	"uspavalia/internal/database"
	"uspavalia/internal/models"
	"uspavalia/internal/services"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

// NewServer registers a Prometheus collector, which panics if done twice in
// one process, so every test in this package shares a single server.
var (
	testOnce   sync.Once
	testServer *httptest.Server
	testDB     = struct{ *Server }{}
	testCfg    *config.Config
)

func setup(t *testing.T) (*httptest.Server, *Server) {
	t.Helper()
	testOnce.Do(func() {
		logrus.SetOutput(io.Discard)
		// Templates are opened relative to the repository root.
		if err := os.Chdir("../.."); err != nil {
			panic(err)
		}
		dir, err := os.MkdirTemp("", "uspavalia-test")
		if err != nil {
			panic(err)
		}
		testCfg = &config.Config{
			DevMode: true,
			Database: config.Database{
				Type: "sqlite",
				Path: filepath.Join(dir, "test.db"),
			},
			Security: config.Security{
				SecretKey: "test-secret-key-test-secret-key!",
				CSRFKey:   "test-csrf-key-test-csrf-key-1234",
			},
		}
		db, err := database.Initialize(testCfg)
		if err != nil {
			panic(err)
		}
		srv := NewServer(testCfg, db)
		testDB.Server = srv
		testServer = httptest.NewServer(srv.Router())
	})
	return testServer, testDB.Server
}

// seed creates a unit, a discipline in it, a professor in it and a scraped
// offering for 2025/1, all with names unique to the calling test.
type seed struct {
	Unit       models.Unit
	OtherUnit  models.Unit
	Discipline models.Discipline
	Professor  models.Professor
	Offering   models.ClassOffering
	User       models.User
}

func seedData(t *testing.T, s *Server, tag string) seed {
	t.Helper()
	var sd seed
	sd.Unit = models.Unit{Name: "Instituto " + tag}
	sd.OtherUnit = models.Unit{Name: "Faculdade " + tag}
	if err := s.db.Create(&sd.Unit).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.db.Create(&sd.OtherUnit).Error; err != nil {
		t.Fatal(err)
	}
	sd.Discipline = models.Discipline{
		Code:   "MAC" + tag,
		Name:   "Cálculo " + tag,
		UnitID: sd.Unit.ID,
	}
	if err := s.db.Create(&sd.Discipline).Error; err != nil {
		t.Fatal(err)
	}
	sd.Professor = models.Professor{Name: "Prof " + tag, UnitID: sd.Unit.ID}
	if err := s.db.Create(&sd.Professor).Error; err != nil {
		t.Fatal(err)
	}
	sd.Offering = models.ClassOffering{
		DisciplineID: sd.Discipline.ID,
		Code:         "2025145",
		StartDate:    "24/02/2025",
		EndDate:      "05/07/2025",
		Type:         "Teórica",
		Schedules:    `[{"dia":"seg","inicio":"08:00","fim":"09:40","professores":["Outra Pessoa"]}]`,
		Vacancies:    "{}",
	}
	if err := s.db.Create(&sd.Offering).Error; err != nil {
		t.Fatal(err)
	}
	sd.User = models.User{EmailHash: "hash-" + tag}
	if err := s.db.Create(&sd.User).Error; err != nil {
		t.Fatal(err)
	}
	return sd
}

// loginCookie forges the session cookie the auth middleware reads.
func loginCookie(t *testing.T, user models.User) *http.Cookie {
	t.Helper()
	store := sessions.NewCookieStore([]byte(testCfg.Security.SecretKey))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	session, _ := store.Get(req, "uspavalia_session")
	session.Values["user_id"] = fmt.Sprintf("%d", user.ID)
	if err := session.Save(req, rec); err != nil {
		t.Fatal(err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no session cookie produced")
	}
	return cookies[0]
}

func noRedirectClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func doForm(t *testing.T, ts *httptest.Server, path string, form url.Values, cookie *http.Cookie) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest("POST", ts.URL+path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := noRedirectClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp, string(body)
}

func doGet(t *testing.T, ts *httptest.Server, path string, cookie *http.Cookie) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest("GET", ts.URL+path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := noRedirectClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp, string(body)
}

func lookup(t *testing.T, ts *httptest.Server, kind, q string, unitID uint) []LookupResult {
	t.Helper()
	path := "/api/lookup/" + kind + "?q=" + url.QueryEscape(q)
	if unitID > 0 {
		path += fmt.Sprintf("&unit_id=%d", unitID)
	}
	resp, body := doGet(t, ts, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("lookup %s: status %d: %s", path, resp.StatusCode, body)
	}
	var results []LookupResult
	if err := json.Unmarshal([]byte(body), &results); err != nil {
		t.Fatalf("lookup %s: bad JSON %q: %v", path, body, err)
	}
	return results
}

func TestLookupFiltersByKindQueryAndUnit(t *testing.T) {
	ts, s := setup(t)
	sd := seedData(t, s, "lk1")
	other := models.Professor{Name: "Prof lk1 other unit", UnitID: sd.OtherUnit.ID}
	if err := s.db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}

	units := lookup(t, ts, "units", "instituto lk1", 0)
	if len(units) != 1 || units[0].ID != sd.Unit.ID {
		t.Fatalf("units lookup = %+v, want the seeded unit", units)
	}

	byCode := lookup(t, ts, "disciplines", "maclk1", 0)
	if len(byCode) != 1 || byCode[0].ID != sd.Discipline.ID || byCode[0].Label != "MAClk1 - Cálculo lk1" {
		t.Fatalf("disciplines by code = %+v", byCode)
	}
	byName := lookup(t, ts, "disciplines", "cálculo lk1", sd.Unit.ID)
	if len(byName) != 1 {
		t.Fatalf("disciplines by name = %+v", byName)
	}
	if got := lookup(t, ts, "disciplines", "cálculo lk1", sd.OtherUnit.ID); len(got) != 0 {
		t.Fatalf("unit filter ignored: %+v", got)
	}

	all := lookup(t, ts, "professors", "prof lk1", 0)
	if len(all) != 2 {
		t.Fatalf("professors without unit filter = %+v, want 2", all)
	}
	filtered := lookup(t, ts, "professors", "prof lk1", sd.Unit.ID)
	if len(filtered) != 1 || filtered[0].ID != sd.Professor.ID {
		t.Fatalf("professors with unit filter = %+v", filtered)
	}

	resp, _ := doGet(t, ts, "/api/lookup/nope?q=x", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown kind status = %d", resp.StatusCode)
	}
}

func TestLookupTreatsWildcardsLiterally(t *testing.T) {
	ts, s := setup(t)
	seedData(t, s, "lk2")
	if got := lookup(t, ts, "professors", "%lk2%", 0); len(got) != 0 {
		t.Fatalf("wildcard query matched %+v", got)
	}
	if got := lookup(t, ts, "professors", "prof_lk2", 0); len(got) != 0 {
		t.Fatalf("underscore query matched %+v", got)
	}
}

func TestAddPageRequiresLogin(t *testing.T) {
	ts, _ := setup(t)
	resp, _ := doGet(t, ts, "/adicionar", nil)
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != "/login" {
		t.Fatalf("status %d location %q, want redirect to /login", resp.StatusCode, resp.Header.Get("Location"))
	}
	resp, _ = doForm(t, ts, "/adicionar", url.Values{}, nil)
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("POST without login status = %d", resp.StatusCode)
	}
}

func TestAddPageRendersWithPrefill(t *testing.T) {
	ts, s := setup(t)
	sd := seedData(t, s, "pg1")
	cookie := loginCookie(t, sd.User)

	resp, body := doGet(t, ts, "/adicionar", cookie)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	for _, want := range []string{`name="discipline_id"`, `name="professor_id"`, `name="year"`, `name="half"`, "/api/lookup/"} {
		if !strings.Contains(body, want) {
			t.Errorf("page missing %q", want)
		}
	}

	_, body = doGet(t, ts, fmt.Sprintf("/adicionar?disciplina_id=%d", sd.Discipline.ID), cookie)
	if !strings.Contains(body, "MACpg1 - Cálculo pg1") || !strings.Contains(body, fmt.Sprintf(`id="discipline_id" value="%d"`, sd.Discipline.ID)) {
		t.Errorf("discipline prefill missing")
	}
	if !strings.Contains(body, "Instituto pg1") {
		t.Errorf("unit not derived from discipline")
	}
	_, body = doGet(t, ts, fmt.Sprintf("/adicionar?professor_id=%d", sd.Professor.ID), cookie)
	if !strings.Contains(body, fmt.Sprintf(`id="professor_id" value="%d"`, sd.Professor.ID)) {
		t.Errorf("professor prefill missing")
	}
	if !strings.Contains(body, "Adicionar Disciplina/Professor</a>") {
		t.Errorf("navbar link missing for logged-in user")
	}
}

func TestAddRejectsFreeTextAndUnknownIDs(t *testing.T) {
	ts, s := setup(t)
	sd := seedData(t, s, "rj1")
	cookie := loginCookie(t, sd.User)
	before := countAll(t, s)

	// Typed names without picking a suggestion: no ids.
	resp, body := doForm(t, ts, "/adicionar", url.Values{
		"discipline_label": {"Uma disciplina inventada"},
		"professor_label":  {"Alguém"},
		"year":             {"2025"},
		"half":             {"1"},
	}, cookie)
	if resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(body, "Selecione uma disciplina") {
		t.Fatalf("free text: status %d", resp.StatusCode)
	}

	// Discipline ok, professor id does not exist.
	resp, body = doForm(t, ts, "/adicionar", url.Values{
		"discipline_id": {fmt.Sprint(sd.Discipline.ID)},
		"professor_id":  {"999999"},
		"year":          {"2025"},
		"half":          {"1"},
	}, cookie)
	if resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(body, "Selecione um professor") {
		t.Fatalf("unknown professor: status %d", resp.StatusCode)
	}

	// Bad semester.
	resp, body = doForm(t, ts, "/adicionar", url.Values{
		"discipline_id": {fmt.Sprint(sd.Discipline.ID)},
		"professor_id":  {fmt.Sprint(sd.Professor.ID)},
		"year":          {"2025"},
		"half":          {"3"},
	}, cookie)
	if resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(body, "semestre válido") {
		t.Fatalf("bad semester: status %d", resp.StatusCode)
	}

	if after := countAll(t, s); after != before {
		t.Fatalf("rows changed on rejected submissions: %v -> %v", before, after)
	}
}

func TestAddCreatesPairAndOfferingFromSemester(t *testing.T) {
	ts, s := setup(t)
	sd := seedData(t, s, "ok1")
	cookie := loginCookie(t, sd.User)

	resp, _ := doForm(t, ts, "/adicionar", url.Values{
		"unit_id":       {fmt.Sprint(sd.Unit.ID)},
		"discipline_id": {fmt.Sprint(sd.Discipline.ID)},
		"professor_id":  {fmt.Sprint(sd.Professor.ID)},
		"year":          {"2025"},
		"half":          {"1"},
	}, cookie)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status %d", resp.StatusCode)
	}

	var cp models.ClassProfessor
	if err := s.db.Where("class_id = ? AND professor_id = ?", sd.Discipline.ID, sd.Professor.ID).First(&cp).Error; err != nil {
		t.Fatalf("pair not created: %v", err)
	}
	if loc := resp.Header.Get("Location"); loc != fmt.Sprintf("/ver/%d", cp.ID) {
		t.Fatalf("redirect = %q", loc)
	}

	var offering models.ClassOffering
	if err := s.db.Where("discipline_id = ? AND code = ?", sd.Discipline.ID, "2025100").First(&offering).Error; err != nil {
		t.Fatalf("offering not created: %v", err)
	}
	// Dates copied from the scraped 2025/1 offering, not the fallback calendar.
	if offering.StartDate != "24/02/2025" || offering.EndDate != "05/07/2025" {
		t.Errorf("dates = %s..%s", offering.StartDate, offering.EndDate)
	}
	if offering.Type != "Teórica" || offering.Notes != services.UserOfferingNote || offering.Vacancies != "{}" {
		t.Errorf("offering fields = %+v", offering)
	}
	var entries []services.ScheduleEntry
	if err := json.Unmarshal([]byte(offering.Schedules), &entries); err != nil {
		t.Fatalf("schedules JSON: %v", err)
	}
	if len(entries) != 1 || len(entries[0].Professors) != 1 || entries[0].Professors[0] != sd.Professor.Name {
		t.Errorf("schedules = %s", offering.Schedules)
	}

	// The pair page lists the semester.
	_, body := doGet(t, ts, fmt.Sprintf("/ver/%d", cp.ID), cookie)
	if !strings.Contains(body, "Semestres registrados") || !strings.Contains(body, "2025/1") {
		t.Errorf("ver page does not list the semester")
	}

	// MatrUSP sees the professor on the (already ended) turma only through
	// the professor list, which is what we set; the client filters past ones.
	// A future semester without scraped data falls back to the calendar.
	resp, _ = doForm(t, ts, "/adicionar", url.Values{
		"discipline_id": {fmt.Sprint(sd.Discipline.ID)},
		"professor_id":  {fmt.Sprint(sd.Professor.ID)},
		"year":          {"2024"},
		"half":          {"2"},
	}, cookie)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("second semester status %d", resp.StatusCode)
	}
	var fallback models.ClassOffering
	if err := s.db.Where("discipline_id = ? AND code = ?", sd.Discipline.ID, "2024200").First(&fallback).Error; err != nil {
		t.Fatalf("fallback offering not created: %v", err)
	}
	if fallback.StartDate != "01/08/2024" || fallback.EndDate != "20/12/2024" {
		t.Errorf("fallback dates = %s..%s", fallback.StartDate, fallback.EndDate)
	}
}

func TestAddIsIdempotentAndAppendsProfessors(t *testing.T) {
	ts, s := setup(t)
	sd := seedData(t, s, "id1")
	cookie := loginCookie(t, sd.User)
	form := url.Values{
		"discipline_id": {fmt.Sprint(sd.Discipline.ID)},
		"professor_id":  {fmt.Sprint(sd.Professor.ID)},
		"year":          {"2025"},
		"half":          {"1"},
	}

	doForm(t, ts, "/adicionar", form, cookie)
	before := countAll(t, s)
	resp, _ := doForm(t, ts, "/adicionar", form, cookie)
	if resp.StatusCode != http.StatusSeeOther || !strings.HasSuffix(resp.Header.Get("Location"), "?existente=1") {
		t.Fatalf("repeat: status %d location %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	if after := countAll(t, s); after != before {
		t.Fatalf("repeat changed rows: %v -> %v", before, after)
	}
	_, body := doGet(t, ts, resp.Header.Get("Location"), cookie)
	if !strings.Contains(body, "já estava registrado") {
		t.Errorf("existing notice missing")
	}

	// A second professor in the same semester joins the same offering.
	second := models.Professor{Name: "Prof id1 segunda", UnitID: sd.Unit.ID}
	if err := s.db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}
	form.Set("professor_id", fmt.Sprint(second.ID))
	resp, _ = doForm(t, ts, "/adicionar", form, cookie)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("second professor status %d", resp.StatusCode)
	}
	var n int64
	s.db.Model(&models.ClassOffering{}).Where("discipline_id = ? AND code = ?", sd.Discipline.ID, "2025100").Count(&n)
	if n != 1 {
		t.Fatalf("offerings for semester = %d, want 1", n)
	}
	var offering models.ClassOffering
	s.db.Where("discipline_id = ? AND code = ?", sd.Discipline.ID, "2025100").First(&offering)
	if !strings.Contains(offering.Schedules, sd.Professor.Name) || !strings.Contains(offering.Schedules, second.Name) {
		t.Errorf("schedules = %s", offering.Schedules)
	}
}

func TestLegacyAddURLRedirects(t *testing.T) {
	ts, _ := setup(t)
	resp, _ := doGet(t, ts, "/?p=add", nil)
	if resp.StatusCode != http.StatusMovedPermanently || resp.Header.Get("Location") != "/adicionar" {
		t.Fatalf("status %d location %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	resp, _ = doGet(t, ts, "/?p=add2&id=7", nil)
	if resp.Header.Get("Location") != "/adicionar?professor_id=7" {
		t.Fatalf("add2 location %q", resp.Header.Get("Location"))
	}
}

type rowCounts struct{ Pairs, Offerings, Disciplines, Professors int64 }

func countAll(t *testing.T, s *Server) rowCounts {
	t.Helper()
	var c rowCounts
	s.db.Model(&models.ClassProfessor{}).Count(&c.Pairs)
	s.db.Model(&models.ClassOffering{}).Count(&c.Offerings)
	s.db.Model(&models.Discipline{}).Count(&c.Disciplines)
	s.db.Model(&models.Professor{}).Count(&c.Professors)
	return c
}
