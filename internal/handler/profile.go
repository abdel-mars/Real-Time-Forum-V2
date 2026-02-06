package handler

import (
	db "forum/internal/db"
	forumerror "forum/internal/error"
	repo "forum/internal/repository"
	utils "forum/internal/utils"
	"net/http"
	"context"
	"time"
)

func ProfilHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(repo.USER_ID_KEY).(int)
	user, err := db.GetUserInfo(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "profile.html", user)
}

func UpddateProfile(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		forumerror.BadRequest(w, r)
		return
	}
	var confMap = make(map[string]any)
	value := r.PathValue("value")
	if r.Context().Value(repo.ERROR_CASE) != nil {
		confMap = r.Context().Value(repo.ERROR_CASE).(map[string]any)
	}

	userId := r.Context().Value(repo.USER_ID_KEY).(int)
	user, err := db.GetUserInfo(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	confMap["Username"] = user.Username

	switch value {
	case "username":
		confMap["username"] = true
		repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "update.html", confMap)
		return
	case "email":
		confMap["email"] = true
		repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "update.html", confMap)
		return
	case "password":
		confMap["password"] = true
		repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "update.html", confMap)
		return
	default:
		forumerror.BadRequest(w, r) // if value is nil the mux will use the root handler
	}
}

func SaveChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		forumerror.BadRequest(w, r)
		return
	}
	switch r.PathValue("value") {
	case "username":
		SaveUsername(w, r)
		return
	case "email":
		SaveEmail(w, r)
		return
	case "password":
		SavePassword(w, r)
		return
	default:
		forumerror.BadRequest(w, r)
		return
	}
}

func SaveUsername(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(repo.USER_ID_KEY).(int)
	new_username := r.FormValue("username")
	password := r.FormValue("current")

	allow, err := db.IsUpdateAllowed(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if !allow {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "You have to wait a 72 hours after your last update \nbefore commiting another"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	if !utils.ValidUsername(new_username) {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Please enter a valid username"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	hash, err := db.GetUserHashById(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if !utils.CheckPassword(password, hash) {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Wrong password"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	dupp, err := db.DupplicatedUsername(new_username)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if dupp {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Username Alredy exists try again"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	err = db.UpdateUsernmae(userId, new_username)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func SaveEmail(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(repo.USER_ID_KEY).(int)
	new_email := r.FormValue("email")
	password := r.FormValue("current")

	allow, err := db.IsUpdateAllowed(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if !allow {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "You have to wait a 72 hours after your last update \nbefore commiting another"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	if !utils.ValidEmail(new_email) {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Invalid email try again"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}
	hash, err := db.GetUserHashById(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if !utils.CheckPassword(password, hash) {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Wrong password"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	dupp, err := db.DupplicatedEmail(new_email)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if dupp {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Please enter a new email"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	err = db.UpdateEmail(userId, new_email)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func SavePassword(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(repo.USER_ID_KEY).(int)
	current := r.FormValue("current")
	new := r.FormValue("new")
	confirm := r.FormValue("confirm")

	allow, err := db.IsUpdateAllowed(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if !allow {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "You have to wait a 72 hours after your last update \nbefore commiting another"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	hash, err := db.GetUserHashById(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	if !utils.ValidPassword(new) {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Invalid password"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}

	if current == new {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "You used an Old password"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}
	if !utils.CheckPassword(current, hash) {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Wrong password"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}
	if new != confirm {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Please Confirm Your password"})
		UpddateProfile(w, r.WithContext(ctx))
		return
	}
	new_hash, err := utils.HashPassword(new)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	err = db.UpdatePassword(userId, new_hash)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func ServeDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		forumerror.BadRequest(w, r)
		return
	}

	var confMap = make(map[string]any)

	if r.Context().Value(repo.ERROR_CASE) != nil {
		confMap = r.Context().Value(repo.ERROR_CASE).(map[string]any)
	}
	userId := r.Context().Value(repo.USER_ID_KEY).(int)
	user, err := db.GetUserInfo(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	confMap["Username"] = user.Username
	repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "delete.html", confMap)
}

// logoutUser handles the logout functionality without importing auth package
func logoutUser(w http.ResponseWriter, r *http.Request, userId int) error {
	// Get session token from cookie
	sessionCookie, err := r.Cookie("session_token")
	if err == nil && sessionCookie.Value != "" {
		// Reset the session in database
		db.ResetUserSession(sessionCookie.Value)
	}

	// Clear the session cookie
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

	return nil
}

func DeleteConfirmation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		forumerror.BadRequest(w, r)
		return
	}

	userId := r.Context().Value(repo.USER_ID_KEY).(int)
	password := r.FormValue("password")
	hash, err := db.GetUserHashById(userId)

	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	if len(password) > repo.PASSWORD_MAX_LEN {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "You exceeded the maximum allowed input"})
		ServeDelete(w, r.WithContext(ctx))
		return
	}

	if !utils.CheckPassword(password, hash) {
		ctx := context.WithValue(r.Context(), repo.ERROR_CASE, map[string]any{"Error": true, "Message": "Wrong password"})
		ServeDelete(w, r.WithContext(ctx))
		return
	}

	// Logout user without calling auth package
	err = logoutUser(w, r, userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	err = db.DeleteUser(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	// Redirect to home page after successful deletion
	http.Redirect(w, r, "/", http.StatusSeeOther)
}