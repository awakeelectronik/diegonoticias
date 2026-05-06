package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/awakeelectronik/diegonoticias/internal/ai"
)

func (h *Handler) generateArticle(w http.ResponseWriter, r *http.Request) {
	if !h.aiLimiter.Allow(time.Now()) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "Cuota diaria de IA agotada"})
		return
	}
	var req ai.GenerateParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	f, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Streaming no soportado"})
		return
	}
	if err := h.aiClient.Stream(r.Context(), req, func(delta string) error {
		_, err := w.Write([]byte("data: " + delta + "\n\n"))
		if err == nil {
			f.Flush()
		}
		return err
	}); err != nil {
		_, _ = w.Write([]byte("event: error\ndata: No se pudo generar el artículo\n\n"))
		f.Flush()
		return
	}
	_, _ = w.Write([]byte("event: done\ndata: {}\n\n"))
	f.Flush()
}

