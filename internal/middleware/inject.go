package middleware

import (
	db "forum/internal/db"
	repo "forum/internal/repository"
	"context"
	"net/http"
)

func InjectUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session_token")
		if err != nil || sessionCookie.Value == "" {
			ctx := context.WithValue(r.Context(), repo.USER_ID_KEY, -1)
			next(w, r.WithContext(ctx))
			return
		}
		userId, exist, err := db.SelectUserSession(sessionCookie.Value)
		if err != nil {
			//forumerror.InternalServerError(w, r, err)
			return
		}
		ctx := r.Context()
		if !exist {
			ctx = context.WithValue(ctx, repo.USER_ID_KEY, -1)
			next(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, repo.USER_ID_KEY, userId)
		usrName, err := db.GetUserNameById(userId)
		if err != nil {
			//forumerror.InternalServerError(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, repo.USER_NAME, usrName)
		next(w, r.WithContext(ctx))

	}
}
