package middleware

import (
	"net/http"
	"strings"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' maxcdn.bootstrapcdn.com cdn.jsdelivr.net cdnjs.cloudflare.com www.google-analytics.com unpkg.com d3js.org; "+
				"style-src 'self' 'unsafe-inline' maxcdn.bootstrapcdn.com cdn.jsdelivr.net cdnjs.cloudflare.com fonts.googleapis.com unpkg.com; "+
				"img-src 'self' data: *.googleusercontent.com; "+
				"font-src 'self' maxcdn.bootstrapcdn.com cdn.jsdelivr.net cdnjs.cloudflare.com fonts.gstatic.com; "+
				"connect-src 'self' accounts.google.com www.google-analytics.com")

		// Remove server header for security
		w.Header().Set("Server", "")

		next.ServeHTTP(w, r)
	})
}

func RedirectToHTTPS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-Proto") == "http" {
			httpsURL := "https://" + r.Host + r.RequestURI
			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func StaticFileHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			// Cache static files for 1 year
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			w.Header().Set("Expires", "Thu, 31 Dec 2025 23:59:59 GMT")

			// Security for static files - prevent execution
			ext := strings.ToLower(r.URL.Path)
			if strings.HasSuffix(ext, ".php") ||
				strings.HasSuffix(ext, ".pl") ||
				strings.HasSuffix(ext, ".py") ||
				strings.HasSuffix(ext, ".jsp") ||
				strings.HasSuffix(ext, ".asp") ||
				strings.HasSuffix(ext, ".sh") ||
				strings.HasSuffix(ext, ".cgi") {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
