package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/awakeelectronik/diegonoticias/internal/ads"
)

func (h *Handler) listAds(w http.ResponseWriter, r *http.Request) {
	items, err := h.adsStore.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "No se pudo cargar publicidad"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"banners": items})
}

func (h *Handler) createAd(w http.ResponseWriter, r *http.Request) {
	var b ads.Banner
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	created, err := h.adsStore.Create(b)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	_ = h.builder.Build()
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateAd(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	var b ads.Banner
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	updated, err := h.adsStore.Update(id, b)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	_ = h.builder.Build()
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteAd(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if err := h.adsStore.Delete(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Banner no encontrado"})
		return
	}
	_ = h.builder.Build()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

