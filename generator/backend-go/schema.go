package backendgo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type contractType struct {
	Name      string
	SchemaRef string
	Root      *structType
	Nested    []structType
	Alias     string
	IsStruct  bool
	Enum      []enumValue
}

type contractField struct {
	Name string
	Type string
	JSON string
}

type structType struct {
	Name   string
	Fields []contractField
}

type enumValue struct {
	Const string
	Value string
}

func loadContract(path string, data []byte, name string) contractType {
	var rawSchema []byte
	if len(data) > 0 {
		rawSchema = data
	} else if path != "" {
		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return contractType{Name: name}
		}
		rawSchema = content
	} else {
		return contractType{Name: name}
	}

	var schema map[string]any
	if err := json.Unmarshal(rawSchema, &schema); err != nil {
		return contractType{Name: name}
	}

	return buildContractFromSchema(name, schema)
}

func loadSchemaLiteral(path string, data []byte) string {
	if len(data) > 0 {
		return strconv.Quote(string(data))
	}
	if path == "" {
		return ""
	}

	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}

	return strconv.Quote(string(content))
}

func buildContractFromSchema(name string, schema map[string]any) contractType {
	if ref, ok := schema["$ref"].(string); ok && ref != "" {
		return contractType{Name: name, Alias: refToGoType(ref)}
	}

	typ, _ := schema["type"].(string)
	switch typ {
	case "object":
		root, nested := buildStructType(name, schema)
		return contractType{Name: name, Root: &root, Nested: nested, IsStruct: true}
	case "array":
		itemType, nested := resolveArrayType(name, "Item", schema["items"])
		return contractType{Name: name, Alias: "[]" + itemType, Nested: nested}
	case "string":
		return contractType{Name: name, Alias: "string", Enum: extractEnum(schema, name)}
	case "integer":
		return contractType{Name: name, Alias: "int", Enum: extractEnum(schema, name)}
	case "number":
		return contractType{Name: name, Alias: "float64", Enum: extractEnum(schema, name)}
	case "boolean":
		return contractType{Name: name, Alias: "bool", Enum: extractEnum(schema, name)}
	}

	return contractType{Name: name, Alias: "map[string]any"}
}

func buildStructType(name string, schema map[string]any) (structType, []structType) {
	props, _ := schema["properties"].(map[string]any)
	if len(props) == 0 {
		return structType{Name: name}, nil
	}

	fields := make([]contractField, 0, len(props))
	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	nested := make([]structType, 0)
	for _, key := range keys {
		subSchema := normalizeSchema(props[key])
		fieldName := pascalCase(key)
		goType, subNested := resolveFieldType(name, fieldName, subSchema)
		nested = append(nested, subNested...)
		fields = append(fields, contractField{
			Name: fieldName,
			Type: goType,
			JSON: key,
		})
	}

	return structType{Name: name, Fields: fields}, nested
}

func resolveFieldType(parentName, fieldName string, schema map[string]any) (string, []structType) {
	if schema == nil {
		return "any", nil
	}

	if ref, ok := schema["$ref"].(string); ok && ref != "" {
		return refToGoType(ref), nil
	}

	typ, _ := schema["type"].(string)
	switch typ {
	case "string":
		return "string", nil
	case "integer":
		return "int", nil
	case "number":
		return "float64", nil
	case "boolean":
		return "bool", nil
	case "object":
		nestedName := parentName + fieldName
		structDecl, nested := buildStructType(nestedName, schema)
		return nestedName, append([]structType{structDecl}, nested...)
	case "array":
		itemType, nested := resolveArrayType(parentName+fieldName, "Item", schema["items"])
		return "[]" + itemType, nested
	}

	if _, ok := schema["anyOf"]; ok {
		return "any", nil
	}

	return "any", nil
}

func resolveArrayType(parentName, suffix string, raw any) (string, []structType) {
	schema := normalizeSchema(raw)
	if schema == nil {
		return "any", nil
	}

	if ref, ok := schema["$ref"].(string); ok && ref != "" {
		return refToGoType(ref), nil
	}

	typ, _ := schema["type"].(string)
	switch typ {
	case "string":
		return "string", nil
	case "integer":
		return "int", nil
	case "number":
		return "float64", nil
	case "boolean":
		return "bool", nil
	case "object":
		nestedName := parentName + suffix
		structDecl, nested := buildStructType(nestedName, schema)
		return nestedName, append([]structType{structDecl}, nested...)
	case "array":
		inner, nested := resolveArrayType(parentName+suffix, "Item", schema["items"])
		return "[]" + inner, nested
	}

	return "any", nil
}

func normalizeSchema(raw any) map[string]any {
	if raw == nil {
		return nil
	}
	if m, ok := raw.(map[string]any); ok {
		return m
	}
	return nil
}

func refToGoType(ref string) string {
	if ref == "" {
		return "any"
	}

	if strings.HasPrefix(ref, "aiwf://") {
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return pascalCase(parts[len(parts)-1])
		}
	}

	base := filepath.Base(ref)
	if dot := strings.Index(base, "."); dot > 0 {
		base = base[:dot]
	}
	return pascalCase(base)
}

func extractEnum(schema map[string]any, typeName string) []enumValue {
	values, _ := schema["enum"].([]any)
	if len(values) == 0 {
		return nil
	}
	enums := make([]enumValue, 0, len(values))
	for _, v := range values {
		if s, ok := v.(string); ok {
			enums = append(enums, enumValue{
				Const: typeName + pascalCase(s),
				Value: s,
			})
		}
	}
	if len(enums) == 0 {
		return nil
	}
	sort.Slice(enums, func(i, j int) bool { return enums[i].Const < enums[j].Const })
	return enums
}
