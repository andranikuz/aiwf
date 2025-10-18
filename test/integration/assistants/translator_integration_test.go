package assistants

import (
	"context"
	"testing"
	"time"

	"github.com/andranikuz/aiwf/test/integration/assistants/generated/translator"
)

func TestTranslator_Integration(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := translator.NewService(openaiClient)

	input := translator.TranslationRequest{
		Text:              "The quick brown fox jumps over the lazy dog",
		SourceLanguage:    "en",
		TargetLanguage:    "es",
		Domain:            "general",
		PreserveFormatting: false,
	}

	result, trace, err := service.Agents().Translator.Run(ctx, input)
	if err != nil {
		t.Fatalf("Translator agent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if trace == nil {
		t.Fatal("Expected trace, got nil")
	}

	t.Logf("✓ Translation completed")
	t.Logf("  Original: %s", input.Text)
	t.Logf("  Translated: %s", result.TranslatedText)
	t.Logf("  Confidence: %.2f", result.ConfidenceScore)

	if len(result.Alternatives) > 0 {
		t.Logf("  Alternative translations (%d):", len(result.Alternatives))
		for i, alt := range result.Alternatives {
			t.Logf("    [%d] %s (preference: %.2f)", i, alt.Text, alt.PreferenceScore)
		}
	}

	if len(result.Notes) > 0 {
		t.Logf("  Translation notes (%d):", len(result.Notes))
		for i, note := range result.Notes {
			t.Logf("    [%d] %s: %q → %q (%s)", i, note.Type, note.OriginalPhrase, note.TranslatedPhrase, note.Explanation)
		}
	}

	t.Logf("  Tokens (in/out): %d/%d", trace.InputTokens, trace.OutputTokens)
}

func TestTranslator_MultipleDomains(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	service := translator.NewService(openaiClient)

	tests := []struct {
		name   string
		text   string
		domain string
		source string
		target string
	}{
		{
			name:   "General English to French",
			text:   "Hello, how are you today?",
			domain: "general",
			source: "en",
			target: "fr",
		},
		{
			name:   "Technical English to German",
			text:   "Initialize the API client and set the authentication token before making requests",
			domain: "technical",
			source: "en",
			target: "de",
		},
		{
			name:   "Legal English to Russian",
			text:   "The party of the first part agrees to indemnify and hold harmless the party of the second part",
			domain: "legal",
			source: "en",
			target: "ru",
		},
		{
			name:   "Medical English to Japanese",
			text:   "The patient presents with acute myocardial infarction and requires immediate intervention",
			domain: "medical",
			source: "en",
			target: "ja",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := translator.TranslationRequest{
				Text:              tt.text,
				SourceLanguage:    tt.source,
				TargetLanguage:    tt.target,
				Domain:            tt.domain,
				PreserveFormatting: true,
			}

			result, trace, err := service.Agents().Translator.Run(ctx, input)
			if err != nil {
				t.Fatalf("Translator failed: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			t.Logf("Domain: %s | %s → %s", tt.domain, tt.source, tt.target)
			t.Logf("  Original: %q", tt.text)
			t.Logf("  Translated: %q", result.TranslatedText)
			t.Logf("  Confidence: %.2f", result.ConfidenceScore)
			t.Logf("  Tokens: %d", trace.InputTokens+trace.OutputTokens)
		})
	}
}

func TestTranslator_LongForm(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := translator.NewService(openaiClient)

	longText := `The development of artificial intelligence has become one of the most important areas of computer science in recent decades.
Machine learning algorithms can now perform tasks that were previously thought to require human intelligence.
Applications range from natural language processing to computer vision, autonomous vehicles, and medical diagnosis.
As AI systems become more powerful and widespread, the ethical implications and societal impacts require careful consideration.`

	input := translator.TranslationRequest{
		Text:              longText,
		SourceLanguage:    "en",
		TargetLanguage:    "es",
		Domain:            "technical",
		PreserveFormatting: true,
	}

	result, trace, err := service.Agents().Translator.Run(ctx, input)
	if err != nil {
		t.Fatalf("Translator failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("✓ Long-form translation completed")
	t.Logf("  Original length: %d characters", len(input.Text))
	t.Logf("  Translated length: %d characters", len(result.TranslatedText))
	t.Logf("  Confidence: %.2f", result.ConfidenceScore)
	t.Logf("  Tokens (in/out): %d/%d", trace.InputTokens, trace.OutputTokens)
}

func TestTranslator_EdgeCases(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	service := translator.NewService(openaiClient)

	tests := []struct {
		name       string
		text       string
		preserveFmt bool
	}{
		{
			name:       "With code snippets",
			text:       "Here is the function: func main() { fmt.Println(\"Hello\") }",
			preserveFmt: true,
		},
		{
			name:       "With URLs",
			text:       "Visit https://example.com for more information about our services",
			preserveFmt: true,
		},
		{
			name:       "With special characters",
			text:       "Price: $99.99 | Rating: ★★★★☆ | Email: test@example.com",
			preserveFmt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := translator.TranslationRequest{
				Text:              tt.text,
				SourceLanguage:    "en",
				TargetLanguage:    "fr",
				Domain:            "general",
				PreserveFormatting: tt.preserveFmt,
			}

			result, _, err := service.Agents().Translator.Run(ctx, input)
			if err != nil {
				t.Fatalf("Translator failed: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			t.Logf("Text: %q", tt.text)
			t.Logf("  → %q", result.TranslatedText)
		})
	}
}
