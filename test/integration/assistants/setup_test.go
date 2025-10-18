package assistants

import (
	"os"
	"testing"

	"github.com/andranikuz/aiwf/providers/openai"
)

var (
	openaiClient *openai.Client
)

func init() {
	// Load API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY environment variable not set")
	}

	// Initialize OpenAI client
	config := openai.ClientConfig{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
	}
	var err error
	openaiClient, err = openai.NewClient(config)
	if err != nil {
		panic("failed to create OpenAI client: " + err.Error())
	}
}

// Helper function to skip tests if API key is not available
func skipIfNoAPIKey(t *testing.T) {
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable not set")
	}
}
