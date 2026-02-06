package handler

import (
	db "forum/internal/db"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

)

// GetChatMessagesHandler handles HTTP requests for chat messages
func GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get current user from session
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currentUserID, hasSession, err := db.SelectUserSession(sessionCookie.Value)
	if err != nil || !hasSession {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameters
	otherUserIDStr := r.URL.Query().Get("other_user_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if otherUserIDStr == "" {
		http.Error(w, "other_user_id required", http.StatusBadRequest)
		return
	}

	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		http.Error(w, "Invalid other_user_id", http.StatusBadRequest)
		return
	}

	limit := 50 // default - increased from 10 to 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	// Get messages using db package function
	messages, err := db.GetChatMessages(currentUserID, otherUserID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	// Mark messages from other user as read using db package function
	db.MarkMessagesAsRead(otherUserID, currentUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"hasMore":  len(messages) == limit,
	})
}

// GetUserByUsernameHandler gets user ID by username for chat
func GetUserByUsernameHandler(w http.ResponseWriter, r *http.Request) {
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
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	userID, email, err := db.GetUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       userID,
		"username": username,
		"email":    email,
	})
}

// GetRecentChatsHandler gets recent chat users for the current user
func GetRecentChatsHandler(w http.ResponseWriter, r *http.Request) {
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
	// Get current user from session
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currentUserID, hasSession, err := db.SelectUserSession(sessionCookie.Value)
	if err != nil || !hasSession {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get recent chat users
	recentUsers, err := db.GetRecentChatUsers(currentUserID, 20)
	if err != nil {
		log.Printf("GetRecentChatsHandler DB error: %v", err)
		// Return empty array instead of error to prevent frontend crashes
		recentUsers = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recent_chats": recentUsers,
	})
}

// GetUnreadCountHandler gets unread message count for current user
func GetUnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	
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
	// Get current user from session
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currentUserID, hasSession, err := db.SelectUserSession(sessionCookie.Value)
	if err != nil || !hasSession {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get unread countS
	unreadCount, err := db.GetUnreadMessageCount(currentUserID)
	if err != nil {
		http.Error(w, "Failed to get unread count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"unread_count": unreadCount,
	})
}


