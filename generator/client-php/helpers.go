package clientphp

import (
	"sort"
	"strings"

	"github.com/andranikuz/aiwf/generator/core"
)

// getTypeName получает имя типа из registry или TypeDef
func (g *Generator) getTypeName(typeName string, typeDef *core.TypeDef) string {
	// Если есть имя типа - используем его
	if typeName != "" {
		return strings.TrimPrefix(typeName, "$")
	}

	// Если есть TypeDef с именем - используем его
	if typeDef != nil && typeDef.Name != "" {
		return typeDef.Name
	}

	return ""
}

// sortFields сортирует имена полей для консистентности
func sortFields(fields []string) {
	sort.Strings(fields)
}
