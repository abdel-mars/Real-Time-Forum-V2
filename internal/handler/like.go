package handler

import (
	repo "forum/internal/repository"
	"encoding/json"
	"net/http"
	"strconv"
	"forum/internal/db"
)

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		if r.Method == http.MethodGet {
			http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method not allowed. Use POST",
		})
		http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
		return
	}
	postId, err := strconv.ParseInt(r.FormValue("post_id"), 10, 0)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid post ID",
		})
		http.Redirect(w, r, "/BadRequest", http.StatusSeeOther)
		return
	}
	IsPostExist, err2 := db.IsPostExist(int(postId))
	if err2 != nil {

		http.Redirect(w, r, "/servererror", http.StatusSeeOther)
		return
	}
	if !IsPostExist {

		http.Redirect(w, r, "/BadRequest", http.StatusSeeOther)
		return
	}
	// Add or remove like
	err = db.AddRemovePostLike(r.Context().Value(repo.USER_ID_KEY).(int), int(postId))
	if err != nil {

		http.Redirect(w, r, "/servererror", http.StatusSeeOther)
		return
	}
	// Get updated post
	post, err := db.GetPostByID(int(postId), r.Context().Value(repo.USER_ID_KEY).(int))
	if err != nil {

		http.Redirect(w, r, "/servererror", http.StatusSeeOther)
		return
	}
	// Success response
	response := map[string]interface{}{
		"success": true,
		"message": "Like updated successfully",
		"postId":  int(postId),
		"data":    post,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
