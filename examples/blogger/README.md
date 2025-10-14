# Blogger Workflow Example

This example demonstrates how to generate a Go SDK from an AIWF workflow and call it using the OpenAI Responses API.

## Files

- `workflow.yaml` — declarative description of assistants and the `blog_post` workflow.
- `schemas/` — JSON Schema definitions for assistant inputs/outputs.
- `sdk/` — generated Go package (regenerate with the command below).
- `main.go` — small CLI that runs the workflow via OpenAI.

## Regenerate SDK

```bash
go run ./cmd/aiwf sdk \
  --file examples/blogger/workflow.yaml \
  --out examples/blogger/sdk \
  --package blog
```

## Run the example

```bash
export OPENAI_API_KEY=... # Responses API key
cd examples/blogger
go run .
```
