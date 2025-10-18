package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

type recordedRequest struct {
	authHeader string
	path       string
	payload    map[string]any
}

func TestNewClientRequiresAPIKey(t *testing.T) {
	if _, err := NewClient(ClientConfig{}); err == nil {
		t.Fatal("expected error when API key is missing")
	}
}

func TestCallJSONSchemaSuccess(t *testing.T) {
	rec := &recordedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.authHeader = r.Header.Get("Authorization")
		rec.path = r.URL.Path

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		rec.payload = payload

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"output": []any{
				map[string]any{
					"content": []any{
						map[string]any{
							"type": "output_text",
							"text": `{"answer":"42"}`,
						},
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 5,
				"total_tokens":      15,
			},
		})
	}))
	defer srv.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:    srv.URL,
		APIKey:     "secret",
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	call := aiwf.ModelCall{
		Model:          "gpt-4.1-mini",
		OutputTypeName: "answer_schema",
		UserPrompt:     "What is the answer?",
		SystemPrompt:   "You are a helpful assistant",
		TypeMetadata: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"answer": map[string]any{"type": "string"},
			},
		},
		MaxTokens:   128,
		Temperature: 0.7,
	}

	raw, usage, err := client.CallJSONSchema(context.Background(), call)
	if err != nil {
		t.Fatalf("CallJSONSchema: %v", err)
	}

	if rec.authHeader != "Bearer secret" {
		t.Errorf("expected auth header, got %q", rec.authHeader)
	}
	if rec.path != "/responses" {
		t.Errorf("expected path /responses, got %s", rec.path)
	}

	if got := string(raw); got != "{\"answer\":\"42\"}" {
		t.Errorf("unexpected raw output: %s", got)
	}

	if usage.Prompt != 10 || usage.Completion != 5 || usage.Total != 15 {
		t.Errorf("unexpected usage: %+v", usage)
	}

	if rec.payload["model"] != "gpt-4.1-mini" {
		t.Errorf("expected model field, got %v", rec.payload["model"])
	}
	t.Logf("payload=%#v", rec.payload)

	input, ok := rec.payload["input"].([]any)
	if !ok {
		t.Fatalf("expected input slice, got %T", rec.payload["input"])
	}
	if len(input) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(input))
	}
	system, _ := input[0].(map[string]any)
	sContent, _ := system["content"].([]any)
	if len(sContent) != 1 {
		t.Fatalf("expected system content block, got %v", system["content"])
	}
	block, _ := sContent[0].(map[string]any)
	if block["text"] != "You are a helpful assistant" {
		t.Fatalf("expected system prompt to propagate, got %v", block["text"])
	}

	textSection, ok := rec.payload["text"].(map[string]any)
	if !ok {
		t.Fatalf("expected text section map, got %T", rec.payload["text"])
	}
	format, _ := textSection["format"].(map[string]any)
	if format["type"] != "json_schema" {
		t.Fatalf("unexpected format type: %v", format["type"])
	}
	jsonSchema, ok := format["json_schema"].(map[string]any)
	if !ok {
		t.Fatalf("expected json_schema map, got %T", format["json_schema"])
	}
	if jsonSchema["strict"] != true {
		t.Fatalf("expected strict json schema, got %v", jsonSchema["strict"])
	}
	if jsonSchema["name"] != "answer_schema" {
		t.Fatalf("unexpected schema name: %v", jsonSchema["name"])
	}
	if _, ok := jsonSchema["schema"].(map[string]any); !ok {
		t.Fatalf("expected schema payload map, got %T", jsonSchema["schema"])
	}

	if v := rec.payload["max_output_tokens"].(float64); v != 128 {
		t.Fatalf("unexpected max_output_tokens: %v", v)
	}
	if v := rec.payload["temperature"].(float64); v != 0.7 {
		t.Fatalf("unexpected temperature: %v", v)
	}
}

func TestCallJSONSchemaRequiresSchema(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"output": map[string]any{},
			"usage":  map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:    srv.URL,
		APIKey:     "secret",
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, _, err = client.CallJSONSchema(context.Background(), aiwf.ModelCall{UserPrompt: "ping"})
	if err == nil {
		t.Fatal("expected error for missing schema ref")
	}
	if !strings.Contains(err.Error(), "output type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStreamingNotImplemented(t *testing.T) {
	client := &Client{}
	ch, _, err := client.CallJSONSchemaStream(context.Background(), aiwf.ModelCall{})
	if err == nil {
		t.Fatal("expected streaming error")
	}
	if _, ok := <-ch; ok {
		t.Fatal("expected closed channel")
	}
}
