package middleware

import (
	"context"
	"net/http"

	// auth "forum/internal/auth"
	db "forum/internal/db"
	forumerror "forum/internal/error"
	repo "forum/internal/repository"
)

func AuthMidleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session_token")

		// Check login; if missing or empty, redirect to login page (root renders login UI)
		if err != nil || sessionCookie.Value == "" {
			// w.wr
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		userId, exist, err := db.SelectUserSession(sessionCookie.Value)
		if err != nil {
			forumerror.InternalServerError(w, r, err)
			return
		}

		if !exist {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), repo.USER_ID_KEY, userId)
		next(w, r.WithContext(ctx))
	}
}
