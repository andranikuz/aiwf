package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/andranikuz/aiwf/test/integration/generated/data_extractor"
)

func TestDataExtractor_Integration(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create service
	service := data_extractor.NewService(openaiClient)

	// Test input
	input := data_extractor.ExtractRequest{
		Text:              "Apple Inc. was founded by Steve Jobs and Steve Wozniak on April 1, 1976. The company is located in Cupertino, California.",
		ExtractionMode:    "entities",
	}

	// Run agent
	result, trace, err := service.Agents().DataExtractor.Run(ctx, input)
	if err != nil {
		t.Fatalf("DataExtractor agent failed: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Check that trace contains token usage
	if trace == nil {
		t.Fatal("Expected trace, got nil")
	}
	if trace.InputTokens == 0 && trace.OutputTokens == 0 {
		t.Logf("Warning: No token usage in trace: %+v", trace)
	}

	// Verify entities were extracted
	if len(result.Entities) == 0 {
		t.Logf("Warning: No entities extracted, result: %+v", result)
	} else {
		t.Logf("✓ Extracted %d entities:", len(result.Entities))
		for i, entity := range result.Entities {
			t.Logf("  [%d] Type: %s, Value: %s, Confidence: %.2f", i, entity.Type, entity.Value, entity.Confidence)
		}
	}

	// Verify metadata
	if result.Metadata == nil {
		t.Fatal("Expected metadata, got nil")
	}
	t.Logf("Metadata: %+v", result.Metadata)
}

func TestDataExtractor_MultipleExtractionModes(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	service := data_extractor.NewService(openaiClient)

	testText := "John Smith, who is an engineer, works at Google in Mountain View. He manages relationships with AWS and Azure teams."

	tests := []struct {
		name        string
		mode        string
		description string
	}{
		{
			name:        "Entities",
			mode:        "entities",
			description: "Extract named entities (people, organizations, locations)",
		},
		{
			name:        "Relationships",
			mode:        "relationships",
			description: "Extract relationships between entities",
		},
		{
			name:        "Full",
			mode:        "full",
			description: "Extract both entities and relationships",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := data_extractor.ExtractRequest{
				Text:           testText,
				ExtractionMode: tt.mode,
			}

			result, trace, err := service.Agents().DataExtractor.Run(ctx, input)
			if err != nil {
				t.Fatalf("DataExtractor failed: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			t.Logf("Mode: %s", tt.mode)
			t.Logf("  Entities: %d", len(result.Entities))
			t.Logf("  Relationships: %d", len(result.Relationships))
			t.Logf("  Tokens (in/out): %d/%d", trace.InputTokens, trace.OutputTokens)
		})
	}
}

func TestDataExtractor_EdgeCases(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := data_extractor.NewService(openaiClient)

	tests := []struct {
		name string
		text string
	}{
		{
			name: "Empty string",
			text: "",
		},
		{
			name: "Simple text",
			text: "Hello world",
		},
		{
			name: "Complex text",
			text: "Dr. Jane Smith and Prof. John Doe from MIT and Stanford published a paper on AI at the 2024 ICML conference in Vienna, Austria.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := data_extractor.ExtractRequest{
				Text:           tt.text,
				ExtractionMode: "full",
			}

			result, _, err := service.Agents().DataExtractor.Run(ctx, input)
			if err != nil {
				t.Fatalf("DataExtractor failed: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			t.Logf("Text: %q → Entities: %d, Relationships: %d",
				tt.text, len(result.Entities), len(result.Relationships))
		})
	}
}
