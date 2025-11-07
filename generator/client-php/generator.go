package clientphp

import (
	"fmt"
	"strings"

	"github.com/andranikuz/aiwf/generator/core"
)

// Generator генерирует PHP HTTP клиент
type Generator struct {
	ir      *core.IR
	baseURL string
}

// New создает новый генератор PHP клиента
func New(ir *core.IR, baseURL string) *Generator {
	return &Generator{
		ir:      ir,
		baseURL: baseURL,
	}
}

// Generate генерирует PHP код
func (g *Generator) Generate() (string, error) {
	var b strings.Builder

	// Header
	b.WriteString("<?php\n")
	b.WriteString("// Auto-generated AIWF HTTP Client\n")
	b.WriteString("// DO NOT EDIT\n\n")

	// Namespace (optional, modern PHP)
	b.WriteString("namespace AIWFClient;\n\n")

	// Generate enums (PHP 8.1+)
	if err := g.generateEnums(&b); err != nil {
		return "", err
	}

	// Generate type classes
	if err := g.generateTypes(&b); err != nil {
		return "", err
	}

	// Generate main client class
	if err := g.generateClient(&b); err != nil {
		return "", err
	}

	return b.String(), nil
}

// generateEnums генерирует enum типы
func (g *Generator) generateEnums(b *strings.Builder) error {
	for typeName, typeDef := range g.ir.Types.Types {
		if typeDef.Kind == core.KindEnum && len(typeDef.Enum) > 0 {
			b.WriteString(fmt.Sprintf("/**\n * Enum for %s\n */\n", typeName))
			b.WriteString(fmt.Sprintf("enum %s: string {\n", typeName))
			for _, value := range typeDef.Enum {
				// Convert to valid PHP constant name
				constName := strings.ToUpper(strings.ReplaceAll(value, "-", "_"))
				b.WriteString(fmt.Sprintf("    case %s = '%s';\n", constName, value))
			}
			b.WriteString("}\n\n")
		}
	}
	return nil
}

// generateTypes генерирует классы для типов
func (g *Generator) generateTypes(b *strings.Builder) error {
	for typeName, typeDef := range g.ir.Types.Types {
		// Skip enums and primitives
		if typeDef.Kind == core.KindEnum || typeDef.Kind != core.KindObject {
			continue
		}

		if typeDef.Properties == nil || len(typeDef.Properties) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("/**\n * %s data structure\n", typeName))
		if typeDef.Description != "" {
			b.WriteString(fmt.Sprintf(" * %s\n", typeDef.Description))
		}
		b.WriteString(" */\n")
		b.WriteString(fmt.Sprintf("class %s {\n", typeName))

		// Get sorted field names for consistent ordering
		fieldNames := make([]string, 0, len(typeDef.Properties))
		for fieldName := range typeDef.Properties {
			fieldNames = append(fieldNames, fieldName)
		}
		// Sort for consistent output
		sortFields(fieldNames)

		// Constructor
		b.WriteString("    public function __construct(\n")
		for i, fieldName := range fieldNames {
			field := typeDef.Properties[fieldName]
			if i > 0 {
				b.WriteString(",\n")
			}

			phpType := g.mapTypeToPHP(field)
			b.WriteString(fmt.Sprintf("        public %s $%s", phpType, fieldName))
		}
		b.WriteString("\n    ) {}\n\n")

		// toArray method for JSON serialization
		b.WriteString("    /**\n")
		b.WriteString("     * Convert to array for JSON encoding\n")
		b.WriteString("     */\n")
		b.WriteString("    public function toArray(): array {\n")
		b.WriteString("        return [\n")
		for _, fieldName := range fieldNames {
			field := typeDef.Properties[fieldName]
			if field.Kind == core.KindArray && field.Items != nil && field.Items.Kind == core.KindRef {
				// Array of objects - need to convert each
				b.WriteString(fmt.Sprintf("            '%s' => array_map(fn($item) => $item->toArray(), $this->%s),\n", fieldName, fieldName))
			} else if field.Kind == core.KindRef {
				// Single object - convert to array
				b.WriteString(fmt.Sprintf("            '%s' => $this->%s->toArray(),\n", fieldName, fieldName))
			} else {
				// Primitive - use as is
				b.WriteString(fmt.Sprintf("            '%s' => $this->%s,\n", fieldName, fieldName))
			}
		}
		b.WriteString("        ];\n")
		b.WriteString("    }\n\n")

		// fromArray static method for deserialization
		b.WriteString("    /**\n")
		b.WriteString("     * Create instance from array\n")
		b.WriteString("     */\n")
		b.WriteString(fmt.Sprintf("    public static function fromArray(array $data): %s {\n", typeName))
		b.WriteString(fmt.Sprintf("        return new %s(\n", typeName))
		for i, fieldName := range fieldNames {
			field := typeDef.Properties[fieldName]
			if i > 0 {
				b.WriteString(",\n")
			}

			// Handle nested objects
			if field.Kind == core.KindRef {
				refType := strings.TrimPrefix(field.Ref, "$")
				b.WriteString(fmt.Sprintf("            %s::fromArray($data['%s'])", refType, fieldName))
			} else if field.Kind == core.KindArray && field.Items != nil && field.Items.Kind == core.KindRef {
				// Handle array of objects
				refType := strings.TrimPrefix(field.Items.Ref, "$")
				b.WriteString(fmt.Sprintf("            array_map(fn($item) => %s::fromArray($item), $data['%s'])", refType, fieldName))
			} else {
				b.WriteString(fmt.Sprintf("            $data['%s']", fieldName))
			}
		}
		b.WriteString("\n        );\n")
		b.WriteString("    }\n")

		b.WriteString("}\n\n")
	}
	return nil
}

// generateClient генерирует основной класс клиента
func (g *Generator) generateClient(b *strings.Builder) error {
	b.WriteString("/**\n")
	b.WriteString(" * AIWF HTTP Client\n")
	b.WriteString(" * \n")
	b.WriteString(fmt.Sprintf(" * Base URL: %s\n", g.baseURL))
	b.WriteString(" */\n")
	b.WriteString("class AIWFClient {\n")
	b.WriteString("    private string $baseURL;\n")
	b.WriteString("    private ?string $apiKey;\n\n")

	// Constructor
	b.WriteString("    public function __construct(\n")
	b.WriteString(fmt.Sprintf("        string $baseURL = '%s',\n", g.baseURL))
	b.WriteString("        ?string $apiKey = null\n")
	b.WriteString("    ) {\n")
	b.WriteString("        $this->baseURL = rtrim($baseURL, '/');\n")
	b.WriteString("        $this->apiKey = $apiKey;\n")
	b.WriteString("    }\n\n")

	// Generate methods for each agent
	for agentName, agent := range g.ir.Assistants {
		if err := g.generateAgentMethod(b, agentName, &agent); err != nil {
			return err
		}
	}

	// Private request method
	b.WriteString("    /**\n")
	b.WriteString("     * Make HTTP request to AIWF server\n")
	b.WriteString("     */\n")
	b.WriteString("    private function request(string $endpoint, array $data): array {\n")
	b.WriteString("        $url = $this->baseURL . $endpoint;\n\n")

	b.WriteString("        $ch = curl_init($url);\n")
	b.WriteString("        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);\n")
	b.WriteString("        curl_setopt($ch, CURLOPT_POST, true);\n")
	b.WriteString("        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));\n\n")

	b.WriteString("        $headers = ['Content-Type: application/json'];\n")
	b.WriteString("        if ($this->apiKey !== null) {\n")
	b.WriteString("            $headers[] = 'X-API-Key: ' . $this->apiKey;\n")
	b.WriteString("        }\n")
	b.WriteString("        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);\n\n")

	b.WriteString("        $response = curl_exec($ch);\n")
	b.WriteString("        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);\n")
	b.WriteString("        curl_close($ch);\n\n")

	b.WriteString("        if ($httpCode !== 200) {\n")
	b.WriteString("            throw new \\Exception(\"HTTP Error $httpCode: $response\");\n")
	b.WriteString("        }\n\n")

	b.WriteString("        $decoded = json_decode($response, true);\n")
	b.WriteString("        if ($decoded === null) {\n")
	b.WriteString("            throw new \\Exception(\"Invalid JSON response: $response\");\n")
	b.WriteString("        }\n\n")

	b.WriteString("        return $decoded;\n")
	b.WriteString("    }\n")

	b.WriteString("}\n")

	return nil
}

// generateAgentMethod генерирует метод для вызова агента
func (g *Generator) generateAgentMethod(b *strings.Builder, agentName string, agent *core.IRAssistant) error {
	methodName := agentName

	// Get input/output types
	inputType := agent.InputType
	outputType := agent.OutputType

	// Check if types are primitives or structs
	inputIsPrimitive := g.isPrimitiveType(inputType)
	outputIsPrimitive := g.isPrimitiveType(outputType)

	// Get type names (resolve from registry if needed)
	inputTypeName := g.getTypeName(agent.InputTypeName, inputType)
	outputTypeName := g.getTypeName(agent.OutputTypeName, outputType)

	b.WriteString("    /**\n")
	b.WriteString(fmt.Sprintf("     * Call %s agent\n", agentName))
	if agent.SystemPrompt != "" {
		// Split system prompt into multiple lines if needed
		lines := strings.Split(agent.SystemPrompt, "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				b.WriteString(fmt.Sprintf("     * %s\n", line))
			}
		}
	}
	b.WriteString("     */\n")

	// Method signature
	if inputIsPrimitive {
		b.WriteString(fmt.Sprintf("    public function %s(string $input)", methodName))
	} else {
		b.WriteString(fmt.Sprintf("    public function %s(%s $request)", methodName, inputTypeName))
	}

	if outputIsPrimitive {
		b.WriteString(": string {\n")
	} else {
		b.WriteString(fmt.Sprintf(": %s {\n", outputTypeName))
	}

	// Method body
	if inputIsPrimitive {
		b.WriteString("        $data = ['input' => $input];\n")
	} else {
		b.WriteString("        $data = $request->toArray();\n")
	}

	b.WriteString(fmt.Sprintf("        $response = $this->request('/agent/%s', $data);\n\n", agentName))

	if outputIsPrimitive {
		b.WriteString("        return $response['output'] ?? $response['result'] ?? '';\n")
	} else {
		b.WriteString(fmt.Sprintf("        return %s::fromArray($response);\n", outputTypeName))
	}

	b.WriteString("    }\n\n")

	return nil
}

// mapTypeToPHP преобразует TypeDef в PHP тип
func (g *Generator) mapTypeToPHP(field *core.TypeDef) string {
	if field == nil {
		return "mixed"
	}

	// Array type
	if field.Kind == core.KindArray {
		return "array"
	}

	// Reference type
	if field.Kind == core.KindRef {
		refType := strings.TrimPrefix(field.Ref, "$")
		return refType
	}

	// Enum type (inline enum - just use string)
	if field.Kind == core.KindEnum {
		if field.Name != "" {
			return field.Name
		}
		return "string" // inline enum без имени
	}

	// Primitive types
	switch field.Kind {
	case core.KindString:
		return "string"
	case core.KindInt:
		return "int"
	case core.KindNumber:
		return "float"
	case core.KindBool:
		return "bool"
	case core.KindAny:
		return "mixed"
	default:
		return "mixed"
	}
}

// isPrimitiveType проверяет является ли тип примитивным
func (g *Generator) isPrimitiveType(typeDef *core.TypeDef) bool {
	if typeDef == nil {
		return false
	}

	primitives := map[core.TypeKind]bool{
		core.KindString: true,
		core.KindInt:    true,
		core.KindNumber: true,
		core.KindBool:   true,
		core.KindAny:    true,
	}
	return primitives[typeDef.Kind]
}
