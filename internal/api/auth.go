package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/awakeelectronik/diegonoticias/internal/auth"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}

	admin, err := auth.LoadAdmin(h.adminFilePath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "No se pudo cargar admin"})
		return
	}

	if req.Username != admin.Username {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Credenciales inválidas"})
		return
	}

	ok, err := auth.ComparePassword(req.Password, admin.PasswordHash)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Credenciales inválidas"})
		return
	}

	token, csrf, expiresAt, err := h.sessions.Create(admin.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "No se pudo crear sesión"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/admin",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	})
	writeJSON(w, http.StatusOK, map[string]string{
		"username":  admin.Username,
		"csrfToken": csrf,
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	username, _ := r.Context().Value(ctxUsernameKey{}).(string)
	csrfToken, _ := r.Context().Value(ctxCSRFKey{}).(string)
	writeJSON(w, http.StatusOK, map[string]string{
		"username":  username,
		"csrfToken": csrfToken,
	})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err == nil {
		h.sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		HttpOnly: true,
		Path:     "/admin",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

