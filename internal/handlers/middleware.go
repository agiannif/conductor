package handlers

import (
	"context"
	"net/http"

	"conductor/internal/models"
)

type contextKey string

const contextKeyUser contextKey = "user"

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		session, err := h.db.GetSession(cookie.Value)
		if err != nil {
			http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Extend sliding window session
		_ = h.db.ExtendSession(cookie.Value)

		user, _, err := h.db.GetUserByID(session.UserID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next(w, r.WithContext(ctx))
	}
}

func currentUser(r *http.Request) models.User {
	u, _ := r.Context().Value(contextKeyUser).(models.User)
	return u
}
