package handler

import (
	"encoding/json"
	"errors"
	 
	db "forum/internal/db"
	forumerror "forum/internal/error"
	repo "forum/internal/repository"
	"net/http"
	"strconv"
)

func FirststateHandler(w http.ResponseWriter, r *http.Request) {
	// Build confMap for response

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

	confMap := make(map[string]any)
	confMap["fileds"] = repo.IT_MAJOR_FIELDS
	// <====== USER AUTH ======>
	userID := -1
	if value := r.Context().Value(repo.USER_ID_KEY); value != nil {
		if id, ok := value.(int); ok {
			userID = id
		}
	}
	username := ""
	if value := r.Context().Value(repo.USER_NAME); value != nil {
		if name, ok := value.(string); ok {
			username = name
		}
	}
	confMap["authenticated"] = userID != -1
	if userID != -1 {
	confMap["userId"] = userID
	confMap["Username"] = username
	if username != "" {
		confMap["Initial"] = username[:1]
	} else {
		confMap["Initial"] = ""
	}
}

	// PAGINATION
	page, err := Pagination(w, r, confMap)
	if err != nil {
		return  
	}
	// Posts
	if err := GetPostsByFilter(w, r, confMap, page); err != nil {
		return  
	}
	//send json response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(confMap)
}

func Pagination(w http.ResponseWriter, r *http.Request, confMap map[string]any) (int, error) {
	query := r.URL.Query()
	page := 1
	filter := query.Get("filter")
	if filter != "" && filter != "Owned" && filter != "Likes" && !repo.IT_MAJOR_FIELDS[filter] {
		forumerror.BadRequest(w, r)
		return -1, errors.New("invalid filter")
	}
	if pageStr := query.Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			forumerror.BadRequest(w, r)
			return -1, err
		}
		page = p
	}
	confMap["CurrentPage"] = page
	confMap["PrintCurrentPage"] = page != 1
	confMap["HasPrev"] = page > 1
	if page > 1 {
		prevQuery := r.URL.Query()
		prevQuery.Set("page", strconv.Itoa(page-1))
		confMap["PrevPage"] = r.URL.Path + "?" + prevQuery.Encode()
	}
	count, err := db.GetPostsCount(query.Get("filter"), r.Context().Value(repo.USER_ID_KEY).(int))
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return -1, err
	}
	// println("posts by count", count)
	// println("posts by calcul", (page - 1) * repo.PAGE_POSTS_QUANTITY)
	if (count <= (page - 1) * repo.PAGE_POSTS_QUANTITY) && page != 1 {
		// forumerror.InternalServerError(w, r, errors.New("invalid page"))
		forumerror.BadRequest(w, r)
		return -1, errors.New("invalid page")
	}
	hasNext := count > page*repo.PAGE_POSTS_QUANTITY
	confMap["HasNext"] = hasNext
	if hasNext {
		nextQuery := r.URL.Query()
		nextQuery.Set("page", strconv.Itoa(page+1))
		confMap["NextPage"] = r.URL.Path + "?" + nextQuery.Encode()
	}
	return page, nil
}

// This Get post by filter !!
func GetPostsByFilter(w http.ResponseWriter, r *http.Request, confMap map[string]any, page int) error {

	query := r.URL.Query()
	userId := r.Context().Value(repo.USER_ID_KEY).(int)

	filter := query.Get("filter")
	if filter == "" {
		data, err := db.GetAllPostsInfo(page, userId)
		if err != nil {
			forumerror.InternalServerError(w, r, err)
			return err
		}
		confMap["Posts"] = data
		return nil
	}
	if filter != "Owned" && filter != "Likes" && !repo.IT_MAJOR_FIELDS[filter] {
		forumerror.BadRequest(w, r)
		return errors.New("resource not found")
	}
	confMap["Filter"] = filter
	switch filter {
	case "Owned":
		if !confMap["Authenticated"].(bool) {
			forumerror.Unauthorized(w, r)
			return errors.New("err")
		}
		data, err := db.Getpostbyowner(userId, page)
		if err != nil {
			forumerror.InternalServerError(w, r, err)
			return err
		}
		confMap["Posts"] = data
	case "Likes":
		if !confMap["Authenticated"].(bool) {
			forumerror.Unauthorized(w, r)
			return errors.New("err")
		}
		data, err := db.Getposbytlikes(userId, page)
		if err != nil {
			forumerror.InternalServerError(w, r, err)
			return err
		}
		confMap["Posts"] = data
	default:
		data, err := db.GePostbycategory(filter, page, userId)
		if err != nil {
			forumerror.InternalServerError(w, r, err)
			return err
		}
		confMap["Posts"] = data
	}
	return nil
}
