package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
	"uspavalia/internal/models"

	"github.com/sirupsen/logrus"
)

func (s *Server) handleMatrusp(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/matrusp/", http.StatusMovedPermanently)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","timestamp":"` + fmt.Sprintf("%d", time.Now().Unix()) + `"}`))
}

// XML sitemap structures following Google's best practices
type SitemapURL struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
	LastMod string   `xml:"lastmod,omitempty"`
}

type SitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

func (s *Server) handleSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	baseURL := fmt.Sprintf("http://%s", r.Host)
	if r.TLS != nil {
		baseURL = fmt.Sprintf("https://%s", r.Host)
	}

	sitemap := &SitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	// Add main pages
	mainPages := []string{
		"/",
		"/sobre",
		"/10melhores",
		"/destaques",
		"/search",
	}

	for _, page := range mainPages {
		sitemap.URLs = append(sitemap.URLs, SitemapURL{
			Loc: baseURL + page,
		})
	}

	// Add all class-professor pages (equivalent to ver pages)
	var classProfessors []models.ClassProfessor
	if err := s.db.Find(&classProfessors).Error; err != nil {
		logrus.Printf("Error fetching class professors for sitemap: %v", err)
	} else {
		for _, cp := range classProfessors {
			sitemap.URLs = append(sitemap.URLs, SitemapURL{
				Loc: fmt.Sprintf("%s/ver/%d", baseURL, cp.ID),
			})
		}
	}

	// Add discipline pages
	var disciplines []models.Discipline
	if err := s.db.Find(&disciplines).Error; err != nil {
		logrus.Printf("Error fetching disciplines for sitemap: %v", err)
	} else {
		for _, discipline := range disciplines {
			sitemap.URLs = append(sitemap.URLs, SitemapURL{
				Loc: fmt.Sprintf("%s/disciplina/%d", baseURL, discipline.ID),
			})
		}
	}

	// Add professor pages
	var professors []models.Professor
	if err := s.db.Find(&professors).Error; err != nil {
		logrus.Printf("Error fetching professors for sitemap: %v", err)
	} else {
		for _, professor := range professors {
			sitemap.URLs = append(sitemap.URLs, SitemapURL{
				Loc: fmt.Sprintf("%s/professor/%d", baseURL, professor.ID),
			})
		}
	}

	// Generate XML
	xmlData, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		logrus.Printf("Error generating sitemap XML: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Write XML declaration and content
	w.Write([]byte(xml.Header))
	w.Write(xmlData)
}
