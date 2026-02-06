
package handler

import (
	auth "forum/internal/auth"
	db "forum/internal/db"
	"encoding/json"
	"net/http"
)

// OnlineUsersHandler returns list of online users for chat selection
func OnlineUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Method not allowed. Only GET is supported.",
		})
		http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
		return
	}

	accept := r.Header.Get("Accept")
	if containsHTML(accept) {
		// If browser wants HTML, redirect to unauthorized
		http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get current user from session to exclude them from the list
	var currentUsername string
	if sessionCookie, err := r.Cookie("session_token"); err == nil {
		if userId, hasSession, err := db.SelectUserSession(sessionCookie.Value); err == nil && hasSession {
			if username, err := db.GetUserNameById(userId); err == nil {
				currentUsername = username
			}
		}
	}

	// Build list of online users (excluding current user)
	onlineUsersList := auth.GetOnlineUsersList()
	var onlineUsers []string
	for _, username := range onlineUsersList {
		if username != currentUsername {
			onlineUsers = append(onlineUsers, username)
		}
	}

	response := map[string]interface{}{
		"users": onlineUsers,
	}

	json.NewEncoder(w).Encode(response)
}

// AllUsersHandler returns list of all registered usernames
func AllUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Method not allowed. Only GET is supported.",
		})
		http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
		return
	}

	accept := r.Header.Get("Accept")
	if containsHTML(accept) {
		http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	users, err := db.GetAllUsernames()
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"users": users})
}
