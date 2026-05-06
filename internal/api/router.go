package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/awakeelectronik/diegonoticias/internal/ai"
	"github.com/awakeelectronik/diegonoticias/internal/ads"
	"github.com/awakeelectronik/diegonoticias/internal/articles"
	"github.com/awakeelectronik/diegonoticias/internal/auth"
	"github.com/awakeelectronik/diegonoticias/internal/builder"
	"github.com/awakeelectronik/diegonoticias/internal/config"
	"github.com/awakeelectronik/diegonoticias/internal/images"
	"github.com/awakeelectronik/diegonoticias/internal/ratelimit"
	"github.com/awakeelectronik/diegonoticias/internal/settings"
)

type Handler struct {
	adminFilePath string
	sessions      *auth.SessionManager
	adminDistDir  string
	sitePublicDir string
	builder       *builder.Builder
	articleStore  *articles.Store
	settingsStore *settings.Store
	adsStore      *ads.Store
	imagePipeline *images.Pipeline
	uploadsDir    string
	aiClient      *ai.Client
	aiLimiter     *ratelimit.DailyLimiter
}

func New(cfg config.Config) *Handler {
	siteDir := cfg.SiteDir
	if siteDir == "" {
		siteDir = "./site"
	}
	maxPerDay := 100
	if raw := strings.TrimSpace(os.Getenv("GROQ_MAX_PER_DAY")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			maxPerDay = n
		}
	}
	return &Handler{
		adminFilePath: filepath.Join(cfg.DataDir, "admin.json"),
		sessions:      auth.NewSessionManager(auth.SessionTTL()),
		adminDistDir:  filepath.Join("web", "admin", "dist"),
		sitePublicDir: filepath.Join(siteDir, "public"),
		builder:       builder.New(siteDir, cfg.DataDir, cfg.HugoBin, cfg.PagefindBin),
		articleStore:  articles.NewStore(filepath.Join(siteDir, "content", "articulos")),
		settingsStore: settings.New(filepath.Join(cfg.DataDir, "settings.json")),
		adsStore:      ads.New(filepath.Join(cfg.DataDir, "ads.json")),
		imagePipeline: images.NewPipeline(images.Config{UploadsRoot: cfg.UploadsDir}),
		uploadsDir:    cfg.UploadsDir,
		aiClient:      ai.New(),
		aiLimiter:     ratelimit.NewDailyLimiter(maxPerDay),
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /admin/api/login", h.login)
	mux.HandleFunc("GET /admin/api/me", h.authRequired(h.me))
	mux.HandleFunc("POST /admin/api/logout", h.csrfRequired(h.logout))
	mux.HandleFunc("GET /admin/api/articulos", h.authRequired(h.listArticles))
	mux.HandleFunc("GET /admin/api/articulos/{slug}", h.authRequired(h.getArticle))
	mux.HandleFunc("POST /admin/api/articulos", h.csrfRequired(h.createArticle))
	mux.HandleFunc("PUT /admin/api/articulos/{slug}", h.csrfRequired(h.updateArticle))
	mux.HandleFunc("DELETE /admin/api/articulos/{slug}", h.csrfRequired(h.deleteArticle))
	mux.HandleFunc("POST /admin/api/articulos/generar", h.csrfRequired(h.generateArticle))
	mux.HandleFunc("POST /admin/api/imagenes", h.csrfRequired(h.uploadImage))
	mux.HandleFunc("GET /admin/api/ajustes", h.authRequired(h.getSettings))
	mux.HandleFunc("PUT /admin/api/ajustes", h.csrfRequired(h.updateSettings))
	mux.HandleFunc("GET /admin/api/publicidad", h.authRequired(h.listAds))
	mux.HandleFunc("POST /admin/api/publicidad", h.csrfRequired(h.createAd))
	mux.HandleFunc("PUT /admin/api/publicidad/{id}", h.csrfRequired(h.updateAd))
	mux.HandleFunc("DELETE /admin/api/publicidad/{id}", h.csrfRequired(h.deleteAd))

	mux.Handle("GET /admin/", h.adminSPAHandler())
	mux.Handle("GET /images/", h.imagesHandler())
	mux.Handle("GET /", h.publicSiteHandler())
	mux.Handle("GET /articulos/", h.publicSiteHandler())

	return securityHeaders(loggingMiddleware(mux))
}

func (h *Handler) BuildInitialIfNeeded() error {
	if _, err := os.Stat(filepath.Join(h.sitePublicDir, "index.html")); err == nil {
		return nil
	}
	res := h.builder.Build()
	if res.Status == "error" {
		return errors.New(res.Error)
	}
	return nil
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

func (h *Handler) publicSiteHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := filepath.Clean("/" + r.URL.Path)
		target := filepath.Join(h.sitePublicDir, rel)
		if fi, err := os.Stat(target); err == nil && fi.IsDir() {
			target = filepath.Join(target, "index.html")
		}
		if fi, err := os.Stat(target); err == nil && !fi.IsDir() {
			if strings.HasSuffix(target, ".html") {
				w.Header().Set("Cache-Control", "no-cache")
			} else {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.ServeFile(w, r, target)
			return
		}
		http.NotFound(w, r)
	})
}

func (h *Handler) imagesHandler() http.Handler {
	fileHandler := http.StripPrefix("/images/", http.FileServer(http.Dir(filepath.Join(h.uploadsDir, "images"))))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		fileHandler.ServeHTTP(w, r)
	})
}

func (h *Handler) imagePipelineMaxBytes() int64 {
	return 8 * 1024 * 1024
}

