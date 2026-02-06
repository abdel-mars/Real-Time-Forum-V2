package handler

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"
	"forum/internal/db"
	"forum/internal/repository"
	"forum/internal/utils"
)

// helper to always return proper JSON
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func PostPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

		if r.Method == http.MethodGet {
			http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method not allowed. Use POST",
		})
		return
	}
	userID := r.Context().Value(repository.USER_ID_KEY).(int)

	canPost, err := db.IsUserCanPostToday(userID)
	if err != nil {
		http.Redirect(w, r, "/servererror", http.StatusSeeOther)
		return
	}

	if !canPost {
		writeJSON(w, http.StatusTooManyRequests,
			map[string]string{"error": "You can only post once per day"})
		return
	}

	title := strings.TrimSpace(html.EscapeString(r.FormValue("title")))
	content := strings.TrimSpace(html.EscapeString(r.FormValue("content")))
	if title == "" || content == "" {
		writeJSON(w, http.StatusBadRequest,
			map[string]string{"error": "Title and content are required"})
		return
	}

	if !utils.ValidPost(content) || !utils.ValidPostTitle(title) {
		writeJSON(w, http.StatusBadRequest,
			map[string]string{"error": "Invalid title or content"})
		return
	}

	categories := r.Form["Categories"]
	for _, c := range categories {
		if !repository.IT_MAJOR_FIELDS[c] {
			writeJSON(w, http.StatusBadRequest,
				map[string]string{"error": "Invalid category"})
			return
		}
	}

	postID, err := db.AddNewPost(userID, title, content)
	if err != nil {
		http.Redirect(w, r, "/servererror", http.StatusSeeOther)
		return
	}

	if err := db.MapPostWithCategories(postID, categories); err != nil {
		http.Redirect(w, r, "/servererror", http.StatusSeeOther)
		return
	}

	// final success JSON ...
	writeJSON(w, http.StatusCreated, map[string]any{
		"success":    true,
		"postID":     postID,
		"title":      title,
		"content":    content,
		"categories": categories,
	})
}
