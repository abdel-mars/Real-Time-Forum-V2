package handler

import (
    repo "forum/internal/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
 	"forum/internal/db"
 
	"forum/internal/utils"
)

func CommentHandler(w http.ResponseWriter, r *http.Request) {
    // Only allow POST
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

    // Parse current page
    Current, _ := strconv.Atoi(r.FormValue("currentPage"))

    userId := r.Context().Value(repo.USER_ID_KEY).(int)

    // Check if user can comment today
    IsUserCanCommenttToday, err := db.IsUserCanCommentToday(userId)
    if err != nil {
        http.Redirect(w, r, "/servererror", http.StatusSeeOther)
        return
    }
    if !IsUserCanCommenttToday {
        http.Redirect(w, r, "/TooManyRequests", http.StatusSeeOther)
        return
    }

    // Validate comment content
    if !utils.ValidComment(r.FormValue("comment")) {
        http.Redirect(w, r, "/BadRequest", http.StatusSeeOther)
        return
    }

    // Parse and check post existence
    postId, err := strconv.ParseInt(r.FormValue("post_id"), 10, 0)
    if err != nil {
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
    comment := strings.TrimSpace(r.FormValue("comment"))
    if comment == "" {
        link := fmt.Sprintf("%s#comment", r.Header.Get("Referer"))
        http.Redirect(w, r, link, http.StatusSeeOther)
        return
    }
    err = db.AddNewComment(userId, int(postId), comment)
    if err != nil {
      
        http.Redirect(w, r, "/servererror", http.StatusSeeOther)
        return
    }
    SENDCOMMENT, _, err := db.GetCommentsByPostPaginated(int(postId), Current, userId)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        http.Redirect(w, r, "/servererror", http.StatusSeeOther)
        return
    }
    var response interface{}
    if Current == 1 {
        response = SENDCOMMENT[0]
    } else {
        response = map[string]interface{}{
            "no": true,
        }
    }

    // Send JSON to frontend
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
