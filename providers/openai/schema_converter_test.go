package openai

import (
	"encoding/json"
	"testing"

	"github.com/andranikuz/aiwf/generator/core"
)

func TestSchemaConverterString(t *testing.T) {
	converter := NewSchemaConverter()

	minLen := 1
	maxLen := 100
	td := &core.TypeDef{
		Name:      "Username",
		Kind:      core.KindString,
		MinLength: &minLen,
		MaxLength: &maxLen,
		Format:    "email",
	}

	schema, err := converter.ConvertToJSONSchema(td)
	if err != nil {
		t.Fatalf("ConvertToJSONSchema: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(schema, &result); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	if result["type"] != "string" {
		t.Errorf("expected type string, got %v", result["type"])
	}
	if result["minLength"] != float64(1) {
		t.Errorf("expected minLength 1, got %v", result["minLength"])
	}
	if result["maxLength"] != float64(100) {
		t.Errorf("expected maxLength 100, got %v", result["maxLength"])
	}
	if result["format"] != "email" {
		t.Errorf("expected format email, got %v", result["format"])
	}
}

func TestSchemaConverterObject(t *testing.T) {
	converter := NewSchemaConverter()

	td := &core.TypeDef{
		Name: "User",
		Kind: core.KindObject,
		Properties: map[string]*core.TypeDef{
			"name": {
				Kind: core.KindString,
			},
			"age": {
				Kind: core.KindInt,
			},
			"active": {
				Kind: core.KindBool,
			},
		},
	}

	schema, err := converter.ConvertToJSONSchema(td)
	if err != nil {
		t.Fatalf("ConvertToJSONSchema: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(schema, &result); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	if result["type"] != "object" {
		t.Errorf("expected type object, got %v", result["type"])
	}

	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties map, got %T", result["properties"])
	}

	if len(props) != 3 {
		t.Errorf("expected 3 properties, got %d", len(props))
	}

	nameSchema, ok := props["name"].(map[string]any)
	if !ok || nameSchema["type"] != "string" {
		t.Errorf("expected name to be string type, got %v", props["name"])
	}

	ageSchema, ok := props["age"].(map[string]any)
	if !ok || ageSchema["type"] != "integer" {
		t.Errorf("expected age to be integer type, got %v", props["age"])
	}

	activeSchema, ok := props["active"].(map[string]any)
	if !ok || activeSchema["type"] != "boolean" {
		t.Errorf("expected active to be boolean type, got %v", props["active"])
	}

	required, ok := result["required"].([]any)
	if !ok {
		t.Fatalf("expected required array, got %T", result["required"])
	}
	if len(required) != 3 {
		t.Errorf("expected all 3 fields to be required, got %d", len(required))
	}
}

func TestSchemaConverterArray(t *testing.T) {
	converter := NewSchemaConverter()

	minItems := 1
	maxItems := 10
	td := &core.TypeDef{
		Name: "StringList",
		Kind: core.KindArray,
		Items: &core.TypeDef{
			Kind: core.KindString,
		},
		MinItems: &minItems,
		MaxItems: &maxItems,
	}

	schema, err := converter.ConvertToJSONSchema(td)
	if err != nil {
		t.Fatalf("ConvertToJSONSchema: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(schema, &result); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	if result["type"] != "array" {
		t.Errorf("expected type array, got %v", result["type"])
	}

	items, ok := result["items"].(map[string]any)
	if !ok || items["type"] != "string" {
		t.Errorf("expected items to be string type, got %v", result["items"])
	}

	if result["minItems"] != float64(1) {
		t.Errorf("expected minItems 1, got %v", result["minItems"])
	}
	if result["maxItems"] != float64(10) {
		t.Errorf("expected maxItems 10, got %v", result["maxItems"])
	}
}

func TestSchemaConverterEnum(t *testing.T) {
	converter := NewSchemaConverter()

	td := &core.TypeDef{
		Name: "Status",
		Kind: core.KindEnum,
		Enum: []string{"active", "inactive", "pending"},
	}

	schema, err := converter.ConvertToJSONSchema(td)
	if err != nil {
		t.Fatalf("ConvertToJSONSchema: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(schema, &result); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	if result["type"] != "string" {
		t.Errorf("expected type string, got %v", result["type"])
	}

	enum, ok := result["enum"].([]any)
	if !ok {
		t.Fatalf("expected enum array, got %T", result["enum"])
	}
	if len(enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(enum))
	}
}