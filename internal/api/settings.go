package api

import (
	"encoding/json"
	"net/http"

	"github.com/awakeelectronik/diegonoticias/internal/settings"
)

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.settingsStore.Get()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "No se pudieron cargar ajustes"})
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var s settings.Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	if err := h.settingsStore.Save(s); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "No se pudieron guardar ajustes"})
		return
	}
	writeJSON(w, http.StatusOK, s)
}

