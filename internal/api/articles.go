package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/awakeelectronik/diegonoticias/internal/articles"
)

func (h *Handler) listArticles(w http.ResponseWriter, r *http.Request) {
	items, err := h.articleStore.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "No se pudieron listar artículos"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) getArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	item, err := h.articleStore.Get(slug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Artículo no encontrado"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) createArticle(w http.ResponseWriter, r *http.Request) {
	var a articles.Article
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	if err := h.articleStore.Create(&a); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	_ = h.builder.Build()
	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) updateArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	var a articles.Article
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	if err := h.articleStore.Update(slug, &a); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	_ = h.builder.Build()
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if err := h.articleStore.Delete(slug); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Artículo no encontrado"})
		return
	}
	_ = h.builder.Build()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

