package handler

import (
	repo "forum/internal/repository"
     "encoding/json"
	"net/http"
	_ "github.com/mattn/go-sqlite3"
)

func PostsHandlerApi(w http.ResponseWriter, r *http.Request) {
    confMap := make(map[string]any)

  
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

    
    userId := r.Context().Value(repo.USER_ID_KEY).(int)
    authenticated := userId != -1
    confMap["Authenticated"] = authenticated

    
    page, err := Pagination(w, r, confMap)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        http.Redirect(w, r, "/BadRequest", http.StatusSeeOther)
        return
    }

     
    if err := GetPostsByFilter(w, r, confMap, page); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        http.Redirect(w, r, "/servererror", http.StatusSeeOther)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(confMap); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        http.Redirect(w, r, "/servererror", http.StatusSeeOther)
        return
    }
}

