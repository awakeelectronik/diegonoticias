package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/awakeelectronik/diegonoticias/internal/auth"
	"github.com/awakeelectronik/diegonoticias/internal/config"
)

type Handler struct {
	adminFilePath string
	sessions      *auth.SessionManager
	adminDistDir  string
}

func New(cfg config.Config) *Handler {
	return &Handler{
		adminFilePath: filepath.Join(cfg.DataDir, "admin.json"),
		sessions:      auth.NewSessionManager(auth.SessionTTL()),
		adminDistDir:  filepath.Join("web", "admin", "dist"),
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /admin/api/login", h.login)
	mux.HandleFunc("GET /admin/api/me", h.authRequired(h.me))
	mux.HandleFunc("POST /admin/api/logout", h.csrfRequired(h.logout))
	mux.HandleFunc("GET /admin/api/articulos", h.authRequired(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Hola, diego"})
	}))

	mux.Handle("GET /admin/", h.adminSPAHandler())
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<h1>Diego Noticias</h1>"))
	})

	return securityHeaders(loggingMiddleware(mux))
}

func (h *Handler) adminSPAHandler() http.Handler {
	fs := http.StripPrefix("/admin/", http.FileServer(http.Dir(h.adminDistDir)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/admin")
		if p == "" || p == "/" {
			http.ServeFile(w, r, filepath.Join(h.adminDistDir, "index.html"))
			return
		}
		full := filepath.Join(h.adminDistDir, strings.TrimPrefix(p, "/"))
		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(h.adminDistDir, "index.html"))
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "interest-cohort=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; font-src 'self'; connect-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func SessionMaxAge(ttl time.Duration) int {
	return int(ttl.Seconds())
}

