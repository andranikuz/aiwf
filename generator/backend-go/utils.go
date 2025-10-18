package backendgo

import (
	"strings"
	"unicode"
)

// toPascalCase converts snake_case or kebab-case to PascalCase
func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// pascalCase is an alias for compatibility with existing code
func pascalCase(s string) string {
	return toPascalCase(s)
}

// unexport makes first letter lowercase
func unexport(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}