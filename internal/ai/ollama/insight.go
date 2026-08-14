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

	"github.com/year-recap/internal/recap/insight"
)

type InsightGenerator struct{ *Generator }

func (g InsightGenerator) Generate(ctx context.Context, facts insight.Facts) (insight.Card, error) {
	return g.GenerateInsight(ctx, facts)
}

func (g *Generator) GenerateInsight(ctx context.Context, facts insight.Facts) (insight.Card, error) {
	factsJSON, err := json.Marshal(facts)
	if err != nil {
		return insight.Card{}, fmt.Errorf("marshal insight facts: %w", err)
	}

	schema := insightSchema()
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return insight.Card{}, fmt.Errorf("marshal insight schema: %w", err)
	}

	requestBody := chatRequest{
		Model: g.model,
		Messages: []chatMessage{
			{Role: "system", Content: insightInstructions()},
			{
				Role: "user",
				Content: strings.Join([]string{
					"Ниже privacy-safe агрегированные факты о поведении пользователя за период. Верни JSON строго по переданной схеме.",
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
			Temperature: 0.4,
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return insight.Card{}, fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return insight.Card{}, fmt.Errorf("build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return insight.Card{}, fmt.Errorf("ollama chat request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return insight.Card{}, fmt.Errorf("read ollama response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return insight.Card{}, fmt.Errorf("ollama chat status %d: %s", resp.StatusCode, compactError(body))
	}

	var envelope chatResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return insight.Card{}, fmt.Errorf("decode ollama response envelope: %w", err)
	}
	content := strings.TrimSpace(envelope.Message.Content)
	if content == "" {
		return insight.Card{}, errors.New("ollama response did not contain message.content")
	}

	var card insight.Card
	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&card); err != nil {
		return insight.Card{}, fmt.Errorf("decode structured insight: %w", err)
	}
	if err := insight.ValidateCard(card); err != nil {
		return insight.Card{}, fmt.Errorf("model returned invalid insight card: %w", err)
	}

	return card, nil
}

func insightInstructions() string {
	return strings.Join([]string{
		"Ты аналитик поведения пользователей классифайда.",
		"Тебе дан агрегированный privacy-safe срез активности пользователя за период: только счётчики, ставки и топ-категория, без ID и без сырых событий.",
		"Проанализируй эти цифры и сформулируй краткий вывод о поведении пользователя за период.",
		"Каждое утверждение должно быть прямым следствием переданных чисел; не придумывай события, покупки, суммы в деньгах, даты или намерения, которых нет во входных данных.",
		"Не выдумывай новые числовые значения, отличные от переданных фактов.",
		"Не упоминай внутренние идентификаторы, модели, API или то, что текст создан ИИ.",
		"title — короткий заголовок вывода (до 80 символов).",
		"description — 1-3 предложения с основным выводом о поведении.",
		"highlights — до 5 коротких пунктов с конкретными наблюдениями, каждый пункт должен опираться на конкретную цифру из FACTS.",
		"Тон: нейтральный, аналитический, без оценок финансового положения и без фамильярности.",
	}, " ")
}

func insightSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "description"},
		"properties": map[string]any{
			"title":       map[string]any{"type": "string"},
			"description": map[string]any{"type": "string"},
			"highlights": map[string]any{
				"type":     "array",
				"maxItems": insight.MaxHighlights,
				"items":    map[string]any{"type": "string"},
			},
		},
	}
}
