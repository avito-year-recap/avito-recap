package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/year-recap/internal/recap/narrative"
)

const defaultBaseURL = "http://localhost:11434"

type Config struct {
	Model     string
	BaseURL   string
	Timeout   time.Duration
	KeepAlive string
}

type Generator struct {
	model     string
	baseURL   string
	keepAlive string
	client    *http.Client
}

func New(config Config) (*Generator, error) {
	config.Model = strings.TrimSpace(config.Model)
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.KeepAlive = strings.TrimSpace(config.KeepAlive)

	if config.Model == "" {
		return nil, errors.New("ollama model is required")
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.Timeout <= 0 {
		config.Timeout = 20 * time.Second
	}
	if config.KeepAlive == "" {
		config.KeepAlive = "5m"
	}

	return &Generator{
		model:     config.Model,
		baseURL:   config.BaseURL,
		keepAlive: config.KeepAlive,
		client:    &http.Client{Timeout: config.Timeout},
	}, nil
}

// Check verifies that the local Ollama server is reachable and that the
// configured model has already been pulled. It does not run inference.
func (g *Generator) Check(ctx context.Context) error {
	payload, err := json.Marshal(map[string]string{"model": g.model})
	if err != nil {
		return fmt.Errorf("marshal ollama model check: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/show", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build ollama model check: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama model check: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read ollama model check: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ollama model %q is unavailable (status %d): %s", g.model, resp.StatusCode, compactError(body))
	}

	return nil
}

func (g *Generator) Generate(ctx context.Context, facts narrative.Facts) (narrative.Story, error) {
	factsJSON, err := json.Marshal(facts)
	if err != nil {
		return narrative.Story{}, fmt.Errorf("marshal narrative facts: %w", err)
	}

	schema := narrativeSchema(facts.EditableCardIDs)
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return narrative.Story{}, fmt.Errorf("marshal narrative schema: %w", err)
	}

	requestBody := chatRequest{
		Model: g.model,
		Messages: []chatMessage{
			{Role: "system", Content: narrativeInstructions()},
			{
				Role: "user",
				Content: strings.Join([]string{
					"Ниже privacy-safe факты уже рассчитанного recap. Верни JSON строго по переданной схеме.",
					"Верни ровно по одному description для каждого ID из editableCardIds; не пропускай и не дублируй карточки.",
					"FACTS:", string(factsJSON),
					"JSON_SCHEMA:", string(schemaJSON),
				}, "\n"),
			},
		},
		Stream:    false,
		Think:     false,
		Format:    schema,
		KeepAlive: g.keepAlive,
		Options: chatOptions{
			Temperature: 0.2,
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return narrative.Story{}, fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return narrative.Story{}, fmt.Errorf("build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return narrative.Story{}, fmt.Errorf("ollama chat request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return narrative.Story{}, fmt.Errorf("read ollama response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return narrative.Story{}, fmt.Errorf("ollama chat status %d: %s", resp.StatusCode, compactError(body))
	}

	var envelope chatResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return narrative.Story{}, fmt.Errorf("decode ollama response envelope: %w", err)
	}
	content := strings.TrimSpace(envelope.Message.Content)
	if content == "" {
		return narrative.Story{}, errors.New("ollama response did not contain message.content")
	}

	var story narrative.Story
	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&story); err != nil {
		return narrative.Story{}, fmt.Errorf("decode structured narrative: %w", err)
	}

	return story, nil
}

type chatRequest struct {
	Model     string         `json:"model"`
	Messages  []chatMessage  `json:"messages"`
	Stream    bool           `json:"stream"`
	Think     bool           `json:"think"`
	Format    map[string]any `json:"format"`
	KeepAlive string         `json:"keep_alive,omitempty"`
	Options   chatOptions    `json:"options,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatOptions struct {
	Temperature float64 `json:"temperature"`
}

type chatResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func narrativeInstructions() string {
	return strings.Join([]string{
		"Ты редактор персональных итогов года для классифайда.",
		"Перепиши только descriptions карточек из editableCardIds на естественном русском языке.",
		"Используй только факты из входного JSON и не придумывай числа, события, покупки, намерения или личные качества.",
		"Не меняй бизнес-решения: behavior, achievements и nextAction уже выбраны детерминированным rule engine.",
		"Никогда не переписывай публичную SHARE-карточку: она исключена из editableCardIds и должна оставаться детерминированной.",
		"Не упоминай внутренние пороги, UUID, правила, модели, API или то, что текст создан ИИ.",
		"Не пиши числовые значения: цифры уже отображаются интерфейсом; используй качественные формулировки без новых количественных утверждений.",
		"Тон: тёплый, короткий, без фамильярности и без оценок финансового положения.",
		"Каждое описание — максимум 2 коротких предложения.",
	}, " ")
}

func narrativeSchema(cardIDs []string) map[string]any {
	ids := make([]any, 0, len(cardIDs))
	for _, id := range cardIDs {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}

	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"cards"},
		"properties": map[string]any{
			"cards": map[string]any{
				"type":        "array",
				"minItems":    len(ids),
				"maxItems":    len(ids),
				"uniqueItems": true,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"id", "description"},
					"properties": map[string]any{
						"id":          map[string]any{"type": "string", "enum": ids},
						"description": map[string]any{"type": "string"},
					},
				},
			},
		},
	}
}

func compactError(body []byte) string {
	text := strings.Join(strings.Fields(string(body)), " ")
	const max = 300
	if len(text) > max {
		return text[:max] + "…"
	}
	return text
}
