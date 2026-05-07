package ai

import (
	"bytes"
	"text/template"
)

type PromptData struct {
	RawText         string
	ToneID          string
	ToneDescription string
	TitleHint       string
	HasImage        bool
}

var promptTpl = template.Must(template.New("groq").Parse(`
Eres un editor profesional de noticias en español de Colombia. Recibirás un texto crudo del usuario y debes producir un artículo completo.

REGLAS DURAS:
- Devuelve SOLO un objeto JSON válido, sin texto antes ni después.
- El JSON debe tener exactamente estos campos: title, body, metaDescription, category, imageAlt.
- title: máximo 12 palabras, en español, sin punto final.
- body: OBLIGATORIO entre 180 y 230 palabras (cuenta cada palabra separada por espacios), redactado en 3 o 4 párrafos separados por una línea en blanco. NO menos de 180. Si tu primer borrador es más corto, AMPLÍA con contexto verificable, antecedentes o citas implícitas hasta llegar al rango. Un body corto se considera respuesta inválida.
- metaDescription: entre 140 y 160 caracteres.
- category: una sola palabra en minúsculas, sin tildes ni espacios.
- imageAlt: 8 a 14 palabras. Si no hay imagen, cadena vacía.

ANTES de cerrar el JSON, vuelve a contar las palabras del body. Si son menos de 180, reescríbelo más extenso. No inventes hechos ni fuentes; añade desarrollo, matices y contexto del propio texto crudo.

CONTEXTO:
- Tono solicitado: {{ .ToneID }} ({{ .ToneDescription }})
- ¿Hay imagen subida?: {{ if .HasImage }}sí{{ else }}no{{ end }}
{{- if .TitleHint }}
- Sugerencia de título: "{{ .TitleHint }}"
{{- end }}

TEXTO CRUDO:
"""
{{ .RawText }}
"""
`))

func BuildPrompt(data PromptData) (string, error) {
	var b bytes.Buffer
	if err := promptTpl.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

