package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/awakeelectronik/diegonoticias/internal/ai"
)

const minBodyWords = 150

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
	out, err := h.completeWithLengthCheck(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type generatedArticle struct {
	Title           string `json:"title"`
	Body            string `json:"body"`
	MetaDescription string `json:"metaDescription"`
	Category        string `json:"category"`
	ImageAlt        string `json:"imageAlt"`
}

func (h *Handler) completeWithLengthCheck(ctx context.Context, req ai.GenerateParams) (*generatedArticle, error) {
	out, err := h.callAIOnce(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(strings.Fields(out.Body)) >= minBodyWords {
		return out, nil
	}
	// Reintento único: la primera respuesta vino corta. Avisamos al modelo.
	retry := req
	retry.RawText = req.RawText + "\n\nNOTA INTERNA: tu respuesta anterior tenía un body demasiado corto. Esta vez asegúrate de que el body tenga ENTRE 180 Y 230 PALABRAS, distribuidas en 3 o 4 párrafos."
	out2, err := h.callAIOnce(ctx, retry)
	if err != nil {
		return out, nil
	}
	if len(strings.Fields(out2.Body)) > len(strings.Fields(out.Body)) {
		return out2, nil
	}
	return out, nil
}

func (h *Handler) callAIOnce(ctx context.Context, req ai.GenerateParams) (*generatedArticle, error) {
	raw, err := h.aiClient.Complete(ctx, req)
	if err != nil {
		return nil, errAIUnavailable
	}
	var out generatedArticle
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, errAIInvalid
	}
	return &out, nil
}

var (
	errAIUnavailable = errAIMessage("No se pudo generar el artículo")
	errAIInvalid     = errAIMessage("Respuesta de IA inválida")
)

type errAIMessage string

func (e errAIMessage) Error() string { return string(e) }

