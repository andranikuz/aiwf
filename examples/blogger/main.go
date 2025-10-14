package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	blog "github.com/andranikuz/aiwf/examples/blogger/sdk"
	"github.com/andranikuz/aiwf/providers/openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var is required")
	}

	client, err := openai.NewClient(openai.ClientConfig{APIKey: apiKey})
	if err != nil {
		log.Fatalf("create openai client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	service := blog.NewService(client)

	input := blog.BlogPostInput{
		Topic: "AI workflow automation",
		Tone:  "enthusiastic",
	}

	output, trace, err := service.Workflows().BlogPost().Run(ctx, input)
	if err != nil {
		log.Fatalf("run workflow: %v", err)
	}

	fmt.Printf("\nTitle: %s\n\n%s\n", output.Title, output.Content)
	if trace != nil {
		fmt.Printf("\nTokens: %+v\n", trace.Usage)
	}
}
