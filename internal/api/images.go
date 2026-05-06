package api

import (
	"io"
	"net/http"
	"strings"
)

func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(h.imagePipelineMaxBytes()); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Formulario inválido"})
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Debes adjuntar una imagen"})
		return
	}
	defer file.Close()
	buf, err := io.ReadAll(io.LimitReader(file, h.imagePipelineMaxBytes()+1))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No se pudo leer la imagen"})
		return
	}
	if int64(len(buf)) > h.imagePipelineMaxBytes() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Imagen supera 8MB"})
		return
	}
	basePath, err := h.imagePipeline.ProcessArticleImage(buf)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	alt := strings.TrimSpace(r.FormValue("alt"))
	writeJSON(w, http.StatusOK, map[string]string{
		"basePath": basePath,
		"alt":      alt,
	})
}

