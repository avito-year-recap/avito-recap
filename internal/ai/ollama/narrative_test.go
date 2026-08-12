package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/narrative"
)

func TestGeneratorUsesLocalChatStructuredOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("local Ollama request unexpectedly used authorization: %q", got)
		}

		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request["model"] != "qwen3:4b" {
			t.Fatalf("model = %v", request["model"])
		}
		if request["stream"] != false {
			t.Fatalf("stream = %v, want false", request["stream"])
		}
		if request["think"] != false {
			t.Fatalf("think = %v, want false", request["think"])
		}
		if request["keep_alive"] != "5m" {
			t.Fatalf("keep_alive = %v", request["keep_alive"])
		}

		messages, ok := request["messages"].([]any)
		if !ok || len(messages) != 2 {
			t.Fatalf("messages = %#v", request["messages"])
		}
		user, ok := messages[1].(map[string]any)
		if !ok {
			t.Fatalf("user message = %#v", messages[1])
		}
		content, _ := user["content"].(string)
		if !strings.Contains(content, "RESEARCHER") || strings.Contains(content, "listingId") || strings.Contains(content, "dialogId") {
			t.Fatalf("unexpected safe facts payload: %s", content)
		}

		format, ok := request["format"].(map[string]any)
		if !ok || format["type"] != "object" {
			t.Fatalf("structured output format = %#v", request["format"])
		}
		properties, _ := format["properties"].(map[string]any)
		cardsSchema, _ := properties["cards"].(map[string]any)
		if cardsSchema["minItems"] != float64(1) || cardsSchema["maxItems"] != float64(1) || cardsSchema["uniqueItems"] != true {
			t.Fatalf("cards schema does not require an exact complete set: %#v", cardsSchema)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"qwen3:4b","message":{"role":"assistant","content":"{\"cards\":[{\"id\":\"intro\",\"description\":\"Твой год сложился из множества небольших открытий.\"}]}"},"done":true}`))
	}))
	defer server.Close()

	generator, err := New(Config{Model: "qwen3:4b", BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	story, err := generator.Generate(context.Background(), narrative.Facts{
		Year:            2025,
		Behavior:        narrative.BehaviorFacts{Code: model.BehaviorResearcher, Title: "Исследователь"},
		NextAction:      narrative.ActionFacts{Code: model.ActionOpenFavorites, Title: "Избранное"},
		EditableCardIDs: []string{"intro"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(story.Cards) != 1 || story.Cards[0].ID != "intro" || story.Cards[0].Description == "" {
		t.Fatalf("story = %+v", story)
	}
}

func TestGeneratorCheckUsesShowEndpoint(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/api/show" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "qwen3:4b" {
			t.Fatalf("model = %q", body.Model)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"details":{"parameter_size":"4B"}}`))
	}))
	defer server.Close()

	generator, err := New(Config{Model: "qwen3:4b", BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err := generator.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("show calls = %d", calls.Load())
	}
}

func TestGeneratorReturnsErrorOnProviderFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"model not found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	generator, err := New(Config{Model: "qwen3:4b", BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := generator.Generate(context.Background(), narrative.Facts{EditableCardIDs: []string{"intro"}}); err == nil {
		t.Fatal("expected provider error")
	}
}

func TestNewRequiresModel(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected missing model error")
	}
}
