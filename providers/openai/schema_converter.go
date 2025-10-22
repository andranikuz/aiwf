package openai

import (
	"encoding/json"
	"fmt"

	"github.com/andranikuz/aiwf/generator/core"
)

// SchemaConverter converts TypeDef to JSON Schema for OpenAI API
type SchemaConverter struct{}

// NewSchemaConverter creates a new converter instance
func NewSchemaConverter() *SchemaConverter {
	return &SchemaConverter{}
}

// ConvertToJSONSchema converts TypeDef to JSON Schema format
func (c *SchemaConverter) ConvertToJSONSchema(td *core.TypeDef) (json.RawMessage, error) {
	schema := c.typeDefToSchema(td)
	data, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	return data, nil
}

// ConvertTypeMetadata converts type metadata (can be TypeDef or map) to JSON Schema
func (c *SchemaConverter) ConvertTypeMetadata(metadata any) (json.RawMessage, error) {
	switch v := metadata.(type) {
	case *core.TypeDef:
		return c.ConvertToJSONSchema(v)
	case map[string]any:
		// Ensure additionalProperties is set for objects
		schema := c.ensureAdditionalProperties(v)
		data, err := json.Marshal(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported metadata type: %T", metadata)
	}
}

// ensureAdditionalProperties recursively adds additionalProperties: false to all objects
func (c *SchemaConverter) ensureAdditionalProperties(schema map[string]any) map[string]any {
	if typeStr, ok := schema["type"].(string); ok && typeStr == "object" {
		// Only add if not already present
		if _, hasAdditional := schema["additionalProperties"]; !hasAdditional {
			schema["additionalProperties"] = false
		}
	}

	// Recursively process properties
	if props, ok := schema["properties"].(map[string]any); ok {
		for _, prop := range props {
			if propMap, ok := prop.(map[string]any); ok {
				c.ensureAdditionalProperties(propMap)
			}
		}
	}

	// Recursively process array items
	if items, ok := schema["items"].(map[string]any); ok {
		c.ensureAdditionalProperties(items)
	}

	return schema
}

// typeDefToSchema converts TypeDef to JSON Schema map
func (c *SchemaConverter) typeDefToSchema(td *core.TypeDef) map[string]any {
	schema := make(map[string]any)

	switch td.Kind {
	case core.KindString:
		schema["type"] = "string"
		if td.MinLength != nil {
			schema["minLength"] = *td.MinLength
		}
		if td.MaxLength != nil {
			schema["maxLength"] = *td.MaxLength
		}
		if td.Pattern != "" {
			schema["pattern"] = td.Pattern
		}
		if td.Format != "" {
			// Map format to JSON Schema format
			switch td.Format {
			case "email":
				schema["format"] = "email"
			case "url":
				schema["format"] = "uri"
			case "uuid":
				schema["format"] = "uuid"
			case "datetime":
				schema["format"] = "date-time"
			case "date":
				schema["format"] = "date"
			}
		}

	case core.KindInt:
		schema["type"] = "integer"
		if td.Min != nil {
			schema["minimum"] = int(*td.Min)
		}
		if td.Max != nil {
			schema["maximum"] = int(*td.Max)
		}

	case core.KindNumber:
		schema["type"] = "number"
		if td.Min != nil {
			schema["minimum"] = *td.Min
		}
		if td.Max != nil {
			schema["maximum"] = *td.Max
		}

	case core.KindBool:
		schema["type"] = "boolean"

	case core.KindEnum:
		schema["type"] = "string"
		schema["enum"] = td.Enum

	case core.KindArray:
		schema["type"] = "array"
		if td.Items != nil {
			schema["items"] = c.typeDefToSchema(td.Items)
		}
		if td.MinItems != nil {
			schema["minItems"] = *td.MinItems
		}
		if td.MaxItems != nil {
			schema["maxItems"] = *td.MaxItems
		}

	case core.KindObject:
		schema["type"] = "object"
		if len(td.Properties) > 0 {
			props := make(map[string]any)
			required := []string{}
			for name, prop := range td.Properties {
				props[name] = c.typeDefToSchema(prop)
				// All fields are required by default in our system
				required = append(required, name)
			}
			schema["properties"] = props
			if len(required) > 0 {
				schema["required"] = required
			}
		}
		schema["additionalProperties"] = false

	case core.KindMap:
		schema["type"] = "object"
		if td.ValueType != nil {
			schema["additionalProperties"] = c.typeDefToSchema(td.ValueType)
		}

	case core.KindRef:
		// For references, we need to inline the referenced type
		// In a real implementation, we'd resolve the reference
		schema["$ref"] = "#/definitions/" + td.Ref

	case core.KindAny:
		// Any type - no type restrictions (don't specify type)
		// OpenAI allows any valid JSON value for this field

	case core.KindDatetime:
		schema["type"] = "string"
		schema["format"] = "date-time"

	case core.KindDate:
		schema["type"] = "string"
		schema["format"] = "date"

	case core.KindUUID:
		schema["type"] = "string"
		schema["format"] = "uuid"

	default:
		// Fallback to string
		schema["type"] = "string"
	}

	return schema
}

// ConvertTypesMap converts a map of TypeDefs to a JSON Schema with definitions
func (c *SchemaConverter) ConvertTypesMap(types map[string]*core.TypeDef, rootTypeName string) (json.RawMessage, error) {
	// Build definitions
	definitions := make(map[string]any)
	for name, td := range types {
		definitions[name] = c.typeDefToSchema(td)
	}

	// Build root schema
	rootSchema := map[string]any{
		"$schema": "http://json-schema.org/draft-07/schema#",
	}

	// Add root type reference or inline
	if rootType, ok := types[rootTypeName]; ok {
		// Inline the root type
		for k, v := range c.typeDefToSchema(rootType) {
			rootSchema[k] = v
		}
	} else {
		// Reference to definitions
		rootSchema["$ref"] = "#/definitions/" + rootTypeName
	}

	// Add definitions if there are any references
	if len(definitions) > 0 {
		rootSchema["definitions"] = definitions
	}

	data, err := json.Marshal(rootSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	return data, nil
}
