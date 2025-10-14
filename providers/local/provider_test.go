package local

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

func TestLocalClientDelegates(t *testing.T) {
	var recorded struct {
		path string
		auth string
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded.path = r.URL.Path
		recorded.auth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"output": []any{
				map[string]any{
					"content": []any{
						map[string]any{
							"type": "output_text",
							"text": `{"value":"ok"}`,
						},
					},
				},
			},
			"usage": map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := NewClient(ClientConfig{Endpoint: srv.URL, APIKey: "token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	raw, _, err := client.CallJSONSchema(context.Background(), aiwf.ModelCall{
		Model:           "dummy",
		OutputSchemaRef: "schema",
		OutputSchema:    json.RawMessage(`{"type":"object"}`),
		UserPrompt:      "ping",
	})
	if err != nil {
		t.Fatalf("CallJSONSchema: %v", err)
	}

	if recorded.path != "/responses" {
		t.Fatalf("expected /responses path, got %s", recorded.path)
	}

	if recorded.auth != "Bearer token" {
		t.Fatalf("expected auth header, got %s", recorded.auth)
	}

	if string(raw) != "{\"value\":\"ok\"}" {
		t.Fatalf("unexpected raw output: %s", string(raw))
	}
}

func TestFallbackKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer local-override" {
			t.Fatalf("expected fallback auth header, got %s", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"output": []any{
				map[string]any{
					"content": []any{
						map[string]any{
							"type": "output_text",
							"text": `{"value":"ok"}`,
						},
					},
				},
			},
			"usage": map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := NewClient(ClientConfig{Endpoint: srv.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if _, _, err := client.CallJSONSchema(context.Background(), aiwf.ModelCall{
		OutputSchemaRef: "schema",
		OutputSchema:    json.RawMessage(`{"type":"object"}`),
		UserPrompt:      "ping",
	}); err != nil {
		t.Fatalf("CallJSONSchema: %v", err)
	}
}
