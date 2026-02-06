package auth

import (
	db "forum/internal/db"
	forumerror "forum/internal/error"
	utils "forum/internal/utils"
	
	"encoding/json"
	"net/http"
	"time"

)

func SubmitLogin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return
	}
	username := payload.Username
	password := payload.Password
	exist, err := db.AlreadyExists(username, username)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if (!utils.ValidUsername(username) && !utils.ValidEmail(username)) || !utils.ValidPassword(password) || !exist {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "invalid credentials, try again",
		})
		return
	}
	userId, hash, err := db.GetUserHashByUsername(username)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}
	if !utils.CheckPassword(password, hash) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "wrong password, try again",
		})
		return
	}

	actualUsername, err := db.GetUserNameById(userId)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	session := GenerateToken(32)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session,
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	err = db.UpdateUserSession(userId, session)
	if err != nil {
		forumerror.InternalServerError(w, r, err)
		return
	}

	AddOnlineUser(actualUsername)
	BroadcastUsers()

	json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"username": actualUsername,
		"userId":   userId,
		"token":    session,
	})
}
