package backendgo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type contractType struct {
	Name   string
	Fields []contractField
}

type contractField struct {
	Name string
	Type string
	JSON string
}

func loadContract(path string, name string) contractType {
	if path == "" {
		return contractType{Name: name}
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return contractType{Name: name}
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return contractType{Name: name}
	}

	props, _ := schema["properties"].(map[string]any)
	if len(props) == 0 {
		return contractType{Name: name}
	}

	fields := make([]contractField, 0, len(props))
	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		raw := props[key]
		fieldName := pascalCase(key)
		fieldType := goType(raw)
		fields = append(fields, contractField{
			Name: fieldName,
			Type: fieldType,
			JSON: key,
		})
	}

	return contractType{Name: name, Fields: fields}
}

func loadSchemaLiteral(path string) string {
	if path == "" {
		return ""
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}

	return strconv.Quote(string(data))
}

func goType(raw any) string {
	obj, _ := raw.(map[string]any)
	if obj == nil {
		return "any"
	}

	switch obj["type"].(string) {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		if items, ok := obj["items"].(map[string]any); ok {
			return "[]" + goType(items)
		}
		return "[]any"
	case "object":
		return "map[string]any"
	default:
		if _, ok := obj["anyOf"]; ok {
			return "any"
		}
		if items, ok := obj["items"].(map[string]any); ok {
			elem := goType(items)
			return "[]" + elem
		}
	}

	return "any"
}
