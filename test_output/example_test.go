package aiwfgen

import (
	"context"
	"testing"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

func TestGeneratedSDK(t *testing.T) {
	// Test type creation
	request := UserRequest{
		Text: "Extract entities from this text",
		Mode: "detailed",
	}

	// Test validation
	if err := ValidateUserRequest(&request); err != nil {
		t.Errorf("validation failed: %v", err)
	}

	// Test entity creation
	entity := Entity{
		Name:       "OpenAI",
		Type:       "organization",
		Confidence: 0.95,
	}

	// Test ExtractedData
	data := ExtractedData{
		Entities: []*Entity{&entity},
		Count:    1,
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	// Test validation
	if err := ValidateExtractedData(&data); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

func TestAgentCreation(t *testing.T) {
	// Mock client for testing
	mockClient := &mockModelClient{}

	// Create service
	service := NewService(mockClient)

	// Check agents were created
	if service.Agents() == nil {
		t.Error("Agents should not be nil")
	}

	if service.Agents().DataExtractor == nil {
		t.Error("DataExtractor agent should not be nil")
	}
}

// mockModelClient implements aiwf.ModelClient for testing
type mockModelClient struct{}

func (m *mockModelClient) CallJSONSchema(ctx context.Context, call aiwf.ModelCall) ([]byte, aiwf.Tokens, error) {
	return []byte(`{"entities":[],"count":0,"metadata":{}}`), aiwf.Tokens{}, nil
}

func (m *mockModelClient) CallJSONSchemaStream(ctx context.Context, call aiwf.ModelCall) (<-chan aiwf.StreamChunk, aiwf.Tokens, error) {
	ch := make(chan aiwf.StreamChunk)
	close(ch)
	return ch, aiwf.Tokens{}, nil
}