package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func RequireAuth(store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "uspavalia_session")

			userID, ok := session.Values["user_id"]
			if !ok || userID == nil {
				if isAPIRequest(r) {
					// A redirect would hand an HTML login page to fetch/ajax
					// callers, which then looks like success. Tell them plainly.
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "uspavalia_session")

			if userID, ok := session.Values["user_id"]; ok && userID != nil {
				ctx := context.WithValue(r.Context(), UserIDKey, userID)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isAPIRequest reports whether the request comes from page script rather
// than a browser navigation.
func isAPIRequest(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")
}

func GetUserID(r *http.Request) (string, bool) {
	userID := r.Context().Value(UserIDKey)
	if userID == nil {
		return "", false
	}
	str, ok := userID.(string)
	return str, ok
}
