package auth

import (
	db "forum/internal/db"
	forumerror "forum/internal/error"
	"encoding/json"
	"net/http"
	"time"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		// No valid session, just return JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "No active session",
		})
		return
	}

	sessionToken := sessionCookie.Value

	userId, hasSession, err := db.SelectUserSession(sessionToken)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	// If there is no active session, clear cookie and return 401
	if !hasSession {
		clearCookie := &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		}
		http.SetCookie(w, clearCookie)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "No active session",
		})
		return
	}

	// Reset session in DB for active sessions
	ok, err := db.ResetUserSession(sessionToken)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid session",
		})
		return
	}

	actualUsername, err := db.GetUserNameById(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	RemoveOnlineUser(actualUsername)
	BroadcastUsers()

	// Clear the session cookie on the client side
	clearCookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, clearCookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Logged out successfully",
		"redirect": "/",
	})
}
