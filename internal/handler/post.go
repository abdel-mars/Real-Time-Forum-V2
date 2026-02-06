package handler

import (
	db "forum/internal/db"
	repo "forum/internal/repository"
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
 
	// Only allow GET requests
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
	var resp repo.PostResponse

	//AUTH CHECK
	userID := r.Context().Value(repo.USER_ID_KEY).(int)
	if userID == -1 {
		resp.Authenticated = false
	} else {
		resp.Authenticated = true
		resp.Username = r.Context().Value(repo.USER_NAME).(string)
		if resp.Username != "" {
			resp.Initial = resp.Username[:1]
		}
	}

	// Get post ID from query parameters
	Idpost := r.URL.Query().Get("Id")
	Id, err := strconv.Atoi(Idpost)
	if err != nil {
		//forumerror.BadRequest(w, r)
		http.Redirect(w, r, "/BadRequest", http.StatusSeeOther)
		return
	}

	//FETCH POST 
	post, err := db.GetPostByID(Id, userID)
	if err == sql.ErrNoRows {
		//http.Redirect(w, r, "/", http.StatusSeeOther)
		w.WriteHeader(http.StatusBadRequest)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if err != nil {
		//forumerror.InternalServerError(w, r, err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/servererror", http.StatusSeeOther)
		return
	}
	resp.Post = post

	//pagination
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			w.WriteHeader(http.StatusBadRequest)
			http.Redirect(w, r, "/BadRequest", http.StatusSeeOther)
			return
		}
		page = p
	}
	resp.CurrentPage = page
	resp.HasPrev = page > 1
	if resp.HasPrev {
		prevQuery := r.URL.Query()
		prevQuery.Set("page", strconv.Itoa(page-1))
		resp.PrevPage = r.URL.Path + "?" + prevQuery.Encode()
	}

	//cmnts
	comments, totalComments, err := db.GetCommentsByPostPaginated(Id, page, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/Servererror", http.StatusSeeOther)

		return
	}
	if float64(page) > math.Ceil(float64(totalComments)/10) && totalComments > 0 {
		//forumerror.BadRequest(w, r)
		w.WriteHeader(http.StatusBadRequest)
			http.Redirect(w, r, "/BadRequest", http.StatusSeeOther)
		return
	}
	if totalComments == 0 && page != 1 {
		//forumerror.BadRequest(w, r)
		w.WriteHeader(http.StatusBadRequest)
			http.Redirect(w, r, "/BadRequest", http.StatusSeeOther)
		return
	}
	resp.Comments = comments
	resp.HasNext = totalComments > page*10
	if resp.HasNext {
		nextQuery := r.URL.Query()
		nextQuery.Set("page", strconv.Itoa(page+1))
		resp.NextPage = r.URL.Path + "?" + nextQuery.Encode()
	}
	//mark as read
	resp.Success = true
	//fmt.Println(resp.Comments)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/servererror", http.StatusSeeOther)
	}
}

func containsHTML(accept string) bool {
	accept = strings.ToLower(accept)
	return strings.Contains(accept, "text/html") || strings.Contains(accept, "application/xhtml+xml")
}
 