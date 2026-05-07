package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type GenerateParams struct {
	RawText   string `json:"rawText"`
	Tone      string `json:"tone"`
	TitleHint string `json:"titleHint"`
	HasImage  bool   `json:"hasImage"`
}

type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

func New() *Client {
	model := strings.TrimSpace(os.Getenv("GROQ_MODEL"))
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return &Client{
		apiKey: strings.TrimSpace(os.Getenv("GROQ_API_KEY")),
		model:  model,
		http:   &http.Client{Timeout: 0},
	}
}

func (c *Client) HasKey() bool { return c.apiKey != "" }

// Complete llama a Groq sin streaming y devuelve el contenido JSON del mensaje.
func (c *Client) Complete(ctx context.Context, p GenerateParams) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("GROQ_API_KEY no configurada")
	}
	prompt, err := BuildPrompt(PromptData{
		RawText:         p.RawText,
		ToneID:          p.Tone,
		ToneDescription: toneDescription(p.Tone),
		TitleHint:       p.TitleHint,
		HasImage:        p.HasImage,
	})
	if err != nil {
		return "", err
	}
	body := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": "Responde solo JSON válido."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.7,
		"max_tokens":      2048,
		"stream":          false,
	}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Groq devolvió %s", resp.Status)
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("sin respuesta del modelo")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func toneDescription(tone string) string {
	switch tone {
	case "informativo":
		return "directo, factual, neutral"
	case "profesional":
		return "formal pero accesible"
	case "institucional":
		return "voz oficial"
	case "academico":
		return "preciso y contextual"
	case "cronica":
		return "narrativo y descriptivo"
	case "editorial":
		return "opinión argumentada"
	case "conversacional":
		return "cercano y simple"
	case "pedagogico":
		return "didáctico para no expertos"
	case "dramatico":
		return "tensión y urgencia"
	case "sensacionalista":
		return "impactante sin inventar hechos"
	default:
		return "neutral"
	}
}

