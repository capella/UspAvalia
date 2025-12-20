package handlers

import (
	"net/http"
	"net/url"
)

// handleLegacyURLs handles old PHP-style URLs and redirects them to new routes
// Old URLs used ?p=page&id=123 format, new URLs use /page/123 format
func (s *Server) handleLegacyURLs(w http.ResponseWriter, r *http.Request) {
	// Only handle root path with query parameters
	if r.URL.Path != "/" {
		s.handle404(w, r)
		return
	}

	page := r.URL.Query().Get("p")
	if page == "" {
		// No legacy parameter, show home page
		s.handleHome(w, r)
		return
	}

	id := r.URL.Query().Get("id")

	var redirectURL string

	switch page {
	case "ver":
		if id != "" {
			redirectURL = "/ver/" + id
		}
	case "disciplina":
		if id != "" {
			redirectURL = "/disciplina/" + id
		}
	case "professor", "pesquisa2":
		// pesquisa2 was old search results showing professor
		if id != "" {
			redirectURL = "/professor/" + id
		}
	case "pesquisa", "search":
		// Old search used /?p=pesquisa&pesquisa=term
		searchTerm := r.URL.Query().Get("pesquisa")
		if searchTerm == "" {
			searchTerm = r.URL.Query().Get("search")
		}
		if searchTerm != "" {
			redirectURL = "/search?q=" + url.QueryEscape(searchTerm)
		} else {
			redirectURL = "/search"
		}
	case "sobre":
		redirectURL = "/sobre"
	case "email", "contato":
		redirectURL = "/contato"
	case "login":
		redirectURL = "/login"
	case "logout":
		redirectURL = "/logout"
	case "10melhores", "destaques":
		redirectURL = "/10melhores"
	case "fb-callback":
		// Old Facebook OAuth callback, redirect to Google OAuth
		redirectURL = "/auth/google"
	case "index":
		redirectURL = "/"
	default:
		// Unknown legacy page, show 404
		s.handle404(w, r)
		return
	}

	if redirectURL == "" {
		// No valid redirect found
		s.handle404(w, r)
		return
	}

	// Perform permanent redirect (301) since old URLs should always use new format
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}
