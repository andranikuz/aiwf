package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

func TestNewClientRequiresKey(t *testing.T) {
	if _, err := NewClient(ClientConfig{}); err == nil {
		t.Fatal("expected error without api key")
	}
}

func TestCallJSONSchema(t *testing.T) {
	var recorded struct {
		path    string
		headers http.Header
		body    map[string]any
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded.path = r.URL.Path
		recorded.headers = r.Header.Clone()
		if err := json.NewDecoder(r.Body).Decode(&recorded.body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{
				{
					"type":        "tool_result",
					"tool_result": map[string]any{"answer": "anthropic"},
				},
			},
			"usage": map[string]any{
				"input_tokens":  8,
				"output_tokens": 6,
				"total_tokens":  14,
			},
		})
	}))
	defer srv.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:    srv.URL,
		APIKey:     "key",
		Version:    "test-version",
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	out, usage, err := client.CallJSONSchema(context.Background(), aiwf.ModelCall{
		Model:           "claude-3",
		OutputSchemaRef: "schema",
		UserPrompt:      "hello",
		SystemPrompt:    "system guidance",
	})
	if err != nil {
		t.Fatalf("CallJSONSchema: %v", err)
	}

	if recorded.path != "/messages" {
		t.Fatalf("expected /messages path, got %s", recorded.path)
	}

	if recorded.headers.Get("x-api-key") != "key" {
		t.Fatalf("missing api key header")
	}

	if recorded.headers.Get("anthropic-version") != "test-version" {
		t.Fatalf("missing version header")
	}

	if string(out) != "{\"answer\":\"anthropic\"}" {
		t.Fatalf("unexpected output: %s", string(out))
	}

	if usage.Total != 14 {
		t.Fatalf("unexpected usage: %+v", usage)
	}

	if recorded.body["model"] != "claude-3" {
		t.Fatalf("expected model field, got %v", recorded.body["model"])
	}

	msgs, ok := recorded.body["messages"].([]any)
	if !ok {
		t.Fatalf("expected messages array, got %T", recorded.body["messages"])
	}
	if len(msgs) != 2 {
		t.Fatalf("expected system and user messages, got %d", len(msgs))
	}

	systemMsg, ok := msgs[0].(map[string]any)
	if !ok || systemMsg["role"] != "system" {
		t.Fatalf("expected first message system role, got %v", msgs[0])
	}
	if content, ok := systemMsg["content"].([]any); !ok || len(content) == 0 || content[0].(map[string]any)["text"] != "system guidance" {
		t.Fatalf("expected system prompt in content, got %v", systemMsg["content"])
	}
}

func TestStreamingStub(t *testing.T) {
	client := &Client{}
	ch, _, err := client.CallJSONSchemaStream(context.Background(), aiwf.ModelCall{})
	if err == nil {
		t.Fatal("expected streaming error")
	}
	if _, ok := <-ch; ok {
		t.Fatal("expected closed stream")
	}
}
