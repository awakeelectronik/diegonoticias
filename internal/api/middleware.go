package api

import (
	"context"
	"net/http"

	"github.com/awakeelectronik/diegonoticias/internal/auth"
)

type ctxUsernameKey struct{}
type ctxCSRFKey struct{}

func (h *Handler) authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.SessionCookieName)
		if err != nil || cookie.Value == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
			return
		}
		sess, ok := h.sessions.Get(cookie.Value)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Sesión inválida"})
			return
		}
		ctx := context.WithValue(r.Context(), ctxUsernameKey{}, sess.Username)
		ctx = context.WithValue(ctx, ctxCSRFKey{}, sess.CSRFToken)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) csrfRequired(next http.HandlerFunc) http.HandlerFunc {
	return h.authRequired(func(w http.ResponseWriter, r *http.Request) {
		expected, _ := r.Context().Value(ctxCSRFKey{}).(string)
		provided := r.Header.Get("X-CSRF-Token")
		if !auth.ValidateCSRF(expected, provided) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "CSRF inválido"})
			return
		}
		next(w, r)
	})
}

