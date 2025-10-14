package blog

// NOTE: Ниже показаны заглушки типов. Замените `map[string]any` на структуры,
//      соответствующие JSON Schema, чтобы получить типизированный SDK.
// DraftInput соответствует схеме draft_input.json.
type DraftInput struct {
    Sections []string `json:"sections"`
    Tone string `json:"tone"`
    Topic string `json:"topic"`
}
// DraftOutput соответствует схеме draft_output.json.
type DraftOutput struct {
    Content string `json:"content"`
    Title string `json:"title"`
}


// OutlineInput соответствует схеме outline_input.json.
type OutlineInput struct {
    Tone string `json:"tone"`
    Topic string `json:"topic"`
}
// OutlineOutput соответствует схеме outline_output.json.
type OutlineOutput struct {
    Sections []string `json:"sections"`
    Title string `json:"title"`
}


// BlogPostInput описывает вход workflow.
type BlogPostInput = OutlineInput

// BlogPostOutput описывает выход workflow.
type BlogPostOutput = DraftOutput

