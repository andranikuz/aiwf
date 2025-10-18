package workflows

import (
	"context"
	"testing"
	"time"

	blog_pipeline "github.com/andranikuz/aiwf/test/integration/workflows/generated/blog_pipeline"
)

func TestBlogPipeline_Integration(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	service := blog_pipeline.NewService(openaiClient)

	styleGuide := &blog_pipeline.StyleGuide{
		Tone:        "conversational",
		Voice:       "active",
		Perspective: "first",
	}

	input := blog_pipeline.BlogRequest{
		Topic:           "Getting Started with Go",
		Keywords:        []string{"go", "golang", "programming", "beginners"},
		TargetAudience:  "Software developers new to Go programming",
		WordCountTarget: 1500,
		StyleGuide:      styleGuide,
	}

	// Test if we can create the agents
	agents := service.Agents()
	if agents == nil {
		t.Fatal("Expected agents, got nil")
	}

	t.Logf("✓ Blog pipeline service initialized")
	t.Logf("  Topic: %s", input.Topic)
	t.Logf("  Target length: %d words", input.WordCountTarget)
	t.Logf("  Keywords: %v", input.Keywords)
	t.Logf("  Style: %s (voice: %s, perspective: %s)",
		input.StyleGuide.Tone, input.StyleGuide.Voice, input.StyleGuide.Perspective)
}

func TestBlogPipeline_ResearchAgent(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := blog_pipeline.NewService(openaiClient)

	input := blog_pipeline.ResearchInput{
		Topic: "Machine Learning for Healthcare",
		Depth: "detailed",
	}

	// Test research agent directly if available
	agents := service.Agents()
	if agents.Researcher == nil {
		t.Skip("Researcher agent not available")
	}

	result, trace, err := agents.Researcher.Run(ctx, input)
	if err != nil {
		t.Fatalf("Researcher agent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("✓ Research completed")
	t.Logf("  Key points: %d", len(result.KeyPoints))
	if len(result.KeyPoints) > 0 {
		for i, kp := range result.KeyPoints {
			t.Logf("    [%d] %s", i, kp)
		}
	}

	t.Logf("  Sources: %d", len(result.Sources))
	if len(result.Sources) > 0 {
		for i, src := range result.Sources {
			t.Logf("    [%d] %s (credibility: %.2f)", i, src.Title, src.Credibility)
		}
	}

	t.Logf("  Summary length: %d chars", len(result.Summary))
	t.Logf("  Tokens (in/out): %d/%d", trace.Usage.Prompt, trace.Usage.Completion)
}

func TestBlogPipeline_OutlineAgent(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := blog_pipeline.NewService(openaiClient)

	researchOutput := &blog_pipeline.ResearchOutput{
		KeyPoints: []string{
			"React is a JavaScript library for building user interfaces",
			"Components are reusable UI building blocks",
			"JSX syntax allows writing HTML-like code in JavaScript",
			"State management is key to React applications",
		},
		Sources: []*blog_pipeline.Source{
			{
				Title:       "React Official Documentation",
				URL:         "https://react.dev",
				Credibility: 0.99,
			},
		},
		Summary: "React is a powerful JavaScript library that has revolutionized how developers build web applications...",
	}

	input := blog_pipeline.OutlineInput{
		Topic:        "React for Beginners",
		ResearchData: researchOutput,
	}

	// Test outline agent directly if available
	agents := service.Agents()
	if agents.Outliner == nil {
		t.Skip("Outliner agent not available")
	}

	result, trace, err := agents.Outliner.Run(ctx, input)
	if err != nil {
		t.Fatalf("Outliner agent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("✓ Outline created")
	t.Logf("  Title: %s", result.Title)
	t.Logf("  Sections: %d", len(result.Sections))
	if len(result.Sections) > 0 {
		for i, sec := range result.Sections {
			t.Logf("    [%d] %s (points: %d)", i, sec.Title, len(sec.Points))
		}
	}
	t.Logf("  Estimated words: %d", result.EstimatedWords)
	t.Logf("  Tokens (in/out): %d/%d", trace.Usage.Prompt, trace.Usage.Completion)
}

func TestBlogPipeline_WriterAgent(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := blog_pipeline.NewService(openaiClient)

	styleGuide := &blog_pipeline.StyleGuide{
		Tone:        "technical",
		Voice:       "active",
		Perspective: "first",
	}

	outline := &blog_pipeline.BlogOutline{
		Title: "Getting Started with Docker",
		Sections: []*blog_pipeline.Section{
			{
				ID:    "intro",
				Title: "Introduction to Docker",
				Points: []string{
					"What is containerization",
					"Benefits of Docker",
					"Docker vs Virtual Machines",
				},
			},
			{
				ID:    "setup",
				Title: "Installation and Setup",
				Points: []string{
					"System requirements",
					"Installation steps",
					"Verification",
				},
			},
		},
		EstimatedWords: 2000,
	}

	input := blog_pipeline.WritingInput{
		Outline:    outline,
		StyleGuide: styleGuide,
	}

	// Test writer agent directly if available
	agents := service.Agents()
	if agents.Writer == nil {
		t.Skip("Writer agent not available")
	}

	result, trace, err := agents.Writer.Run(ctx, input)
	if err != nil {
		t.Fatalf("Writer agent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("✓ Draft written")
	t.Logf("  Title: %s", result.Title)
	t.Logf("  Word count: %d", result.WordCount)
	t.Logf("  Content preview: %s...", result.Content[:100])
	t.Logf("  Tokens (in/out): %d/%d", trace.Usage.Prompt, trace.Usage.Completion)
}

func TestBlogPipeline_EditorAgent(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := blog_pipeline.NewService(openaiClient)

	draft := &blog_pipeline.Draft{
		Title:     "Python Best Practices",
		Content:   "Python is a great programming language. You should use it for many things. It has many features that make coding easier.",
		WordCount: 50,
	}

	input := blog_pipeline.EditInput{
		Draft: draft,
	}

	// Test editor agent directly if available
	agents := service.Agents()
	if agents.Editor == nil {
		t.Skip("Editor agent not available")
	}

	result, trace, err := agents.Editor.Run(ctx, input)
	if err != nil {
		t.Fatalf("Editor agent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("✓ Content edited and published")
	t.Logf("  Title: %s", result.Title)
	t.Logf("  Excerpt: %s...", result.Excerpt[:50])
	t.Logf("  Tags: %v", result.Metadata.Tags)
	t.Logf("  Author: %s", result.Metadata.Author.Name)
	t.Logf("  Reading time: %d minutes", result.Metadata.ReadingTimeMinutes)
	t.Logf("  Tokens (in/out): %d/%d", trace.Usage.Prompt, trace.Usage.Completion)
}

func TestBlogPipeline_AllAgents(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := blog_pipeline.NewService(openaiClient)

	t.Logf("✓ Testing all blog pipeline agents")

	agents := service.Agents()

	t.Run("Agents exist", func(t *testing.T) {
		if agents.Researcher != nil {
			t.Log("  ✓ Researcher agent available")
		} else {
			t.Log("  ✗ Researcher agent not available")
		}

		if agents.Outliner != nil {
			t.Log("  ✓ Outliner agent available")
		} else {
			t.Log("  ✗ Outliner agent not available")
		}

		if agents.Writer != nil {
			t.Log("  ✓ Writer agent available")
		} else {
			t.Log("  ✗ Writer agent not available")
		}

		if agents.Editor != nil {
			t.Log("  ✓ Editor agent available")
		} else {
			t.Log("  ✗ Editor agent not available")
		}
	})
}
