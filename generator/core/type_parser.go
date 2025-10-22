package core

import (
	"fmt"
	"strconv"
	"strings"
)

// TypeParser парсит типы из YAML в TypeDef
type TypeParser struct {
	registry *TypeRegistry
}

// NewTypeParser создаёт новый парсер типов
func NewTypeParser() *TypeParser {
	return &TypeParser{
		registry: &TypeRegistry{
			Types:   make(map[string]*TypeDef),
			Imports: make(map[string]*TypeRegistry),
		},
	}
}

// ParseTypes парсит секцию types из YAML
func (p *TypeParser) ParseTypes(types map[string]interface{}) (*TypeRegistry, error) {
	for typeName, typeData := range types {
		typeDef, err := p.parseTypeDefinition(typeName, typeData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse type %s: %w", typeName, err)
		}
		p.registry.Types[typeName] = typeDef
	}
	return p.registry, nil
}

// parseTypeDefinition парсит определение одного типа
func (p *TypeParser) parseTypeDefinition(name string, data interface{}) (*TypeDef, error) {
	// Если это строка, то это простое выражение типа
	if expr, ok := data.(string); ok {
		td, err := ParseTypeExpression(expr)
		if err != nil {
			return nil, err
		}
		td.Name = name
		td.ID = fmt.Sprintf("aiwf://%s", name)
		return td, nil
	}

	// Если это map, то это объект с полями
	if obj, ok := data.(map[interface{}]interface{}); ok {
		return p.parseObjectType(name, obj)
	}

	// Если это map[string]interface{} (альтернативный формат от YAML парсера)
	if obj, ok := data.(map[string]interface{}); ok {
		return p.parseObjectTypeString(name, obj)
	}

	return nil, fmt.Errorf("unexpected type format for %s: %T", name, data)
}

// parseObjectType парсит объектный тип
func (p *TypeParser) parseObjectType(name string, obj map[interface{}]interface{}) (*TypeDef, error) {
	typeDef := &TypeDef{
		Name:       name,
		ID:         fmt.Sprintf("aiwf://%s", name),
		Kind:       KindObject,
		Properties: make(map[string]*TypeDef),
	}

	for key, value := range obj {
		fieldName, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("field name must be string, got %T", key)
		}

		// Check if field is optional (ends with ?)
		isOptional := false
		actualFieldName := fieldName
		if strings.HasSuffix(fieldName, "?") {
			isOptional = true
			actualFieldName = strings.TrimSuffix(fieldName, "?")
		}

		fieldType, err := p.parseFieldType(actualFieldName, value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %s: %w", fieldName, err)
		}
		fieldType.Optional = isOptional
		typeDef.Properties[actualFieldName] = fieldType
	}

	return typeDef, nil
}

// parseObjectTypeString парсит объектный тип (версия для map[string]interface{})
func (p *TypeParser) parseObjectTypeString(name string, obj map[string]interface{}) (*TypeDef, error) {
	typeDef := &TypeDef{
		Name:       name,
		ID:         fmt.Sprintf("aiwf://%s", name),
		Kind:       KindObject,
		Properties: make(map[string]*TypeDef),
	}

	for fieldName, value := range obj {
		// Check if field is optional (ends with ?)
		isOptional := false
		actualFieldName := fieldName
		if strings.HasSuffix(fieldName, "?") {
			isOptional = true
			actualFieldName = strings.TrimSuffix(fieldName, "?")
		}

		fieldType, err := p.parseFieldType(actualFieldName, value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %s: %w", fieldName, err)
		}
		fieldType.Optional = isOptional
		typeDef.Properties[actualFieldName] = fieldType
	}

	return typeDef, nil
}

// parseFieldType парсит тип отдельного поля
func (p *TypeParser) parseFieldType(fieldName string, value interface{}) (*TypeDef, error) {
	// Если это строка - парсим выражение типа
	if expr, ok := value.(string); ok {
		return ParseTypeExpression(expr)
	}

	// Если это вложенный объект
	if obj, ok := value.(map[interface{}]interface{}); ok {
		return p.parseObjectType(fieldName, obj)
	}

	if obj, ok := value.(map[string]interface{}); ok {
		return p.parseObjectTypeString(fieldName, obj)
	}

	// Если это массив (для inline определений полей массива)
	if arr, ok := value.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, fmt.Errorf("empty array definition for field %s", fieldName)
		}
		// Парсим первый элемент как тип элемента массива
		itemType, err := p.parseFieldType(fieldName+"_item", arr[0])
		if err != nil {
			return nil, err
		}
		return &TypeDef{
			Kind:  KindArray,
			Items: itemType,
		}, nil
	}

	return nil, fmt.Errorf("unexpected field type for %s: %T", fieldName, value)
}

// ParseTypeExpressionFull парсит полное выражение типа с ограничениями
func ParseTypeExpressionFull(expr string) (*TypeDef, error) {
	expr = strings.TrimSpace(expr)

	// Проверяем на массив с ограничениями: Type[](min:1, max:10)
	if idx := strings.Index(expr, "[]("); idx > 0 && strings.HasSuffix(expr, ")") {
		baseExpr := expr[:idx]
		constraints := expr[idx+3 : len(expr)-1]

		itemType, err := ParseTypeExpression(baseExpr)
		if err != nil {
			return nil, err
		}

		arrayType := &TypeDef{
			Kind:  KindArray,
			Items: itemType,
		}

		// Парсим ограничения массива
		for _, constraint := range strings.Split(constraints, ",") {
			constraint = strings.TrimSpace(constraint)
			if strings.HasPrefix(constraint, "min:") {
				val, err := strconv.Atoi(strings.TrimPrefix(constraint, "min:"))
				if err != nil {
					return nil, fmt.Errorf("invalid min value: %s", constraint)
				}
				arrayType.MinItems = &val
			} else if strings.HasPrefix(constraint, "max:") {
				val, err := strconv.Atoi(strings.TrimPrefix(constraint, "max:"))
				if err != nil {
					return nil, fmt.Errorf("invalid max value: %s", constraint)
				}
				arrayType.MaxItems = &val
			}
		}

		return arrayType, nil
	}

	// Проверяем на enum и map до проверки на скобки
	if strings.HasPrefix(expr, "enum(") || strings.HasPrefix(expr, "map(") {
		return ParseTypeExpression(expr)
	}

	// Проверяем на тип с ограничениями: string(1..100) или int(0..999)
	if idx := strings.Index(expr, "("); idx > 0 && strings.HasSuffix(expr, ")") {
		baseType := expr[:idx]
		constraints := expr[idx+1 : len(expr)-1]

		td := &TypeDef{}

		switch baseType {
		case "string":
			td.Kind = KindString
			if err := parseStringConstraints(td, constraints); err != nil {
				return nil, err
			}
		case "int":
			td.Kind = KindInt
			if err := parseNumberConstraints(td, constraints); err != nil {
				return nil, err
			}
		case "number":
			td.Kind = KindNumber
			if err := parseNumberConstraints(td, constraints); err != nil {
				return nil, err
			}
		default:
			// Treat as reference with constraints (shouldn't happen for refs, but be safe)
			td = &TypeDef{
				Kind: KindRef,
				Ref:  baseType,
			}
		}

		return td, nil
	}

	// Остальное обрабатывается базовой функцией
	return ParseTypeExpression(expr)
}

// parseStringConstraints парсит ограничения для строк
func parseStringConstraints(td *TypeDef, constraints string) error {
	// Проверяем диапазон длины: 1..100
	if strings.Contains(constraints, "..") {
		parts := strings.Split(constraints, "..")
		if len(parts) != 2 {
			return fmt.Errorf("invalid range format: %s", constraints)
		}

		if parts[0] != "" {
			min, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return fmt.Errorf("invalid min length: %s", parts[0])
			}
			td.MinLength = &min
		}

		if parts[1] != "" {
			max, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return fmt.Errorf("invalid max length: %s", parts[1])
			}
			td.MaxLength = &max
		}
	} else {
		// Это может быть формат: email, url, phone
		td.Format = constraints
	}

	return nil
}

// parseNumberConstraints парсит ограничения для чисел
func parseNumberConstraints(td *TypeDef, constraints string) error {
	// Проверяем диапазон: 0..100 или 1..
	if strings.Contains(constraints, "..") {
		parts := strings.Split(constraints, "..")
		if len(parts) != 2 {
			return fmt.Errorf("invalid range format: %s", constraints)
		}

		if parts[0] != "" {
			min, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if err != nil {
				return fmt.Errorf("invalid min value: %s", parts[0])
			}
			td.Min = &min
		}

		if parts[1] != "" {
			max, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return fmt.Errorf("invalid max value: %s", parts[1])
			}
			td.Max = &max
		}
	} else {
		// Одно число - точное значение?
		return fmt.Errorf("single number constraint not supported: %s", constraints)
	}

	return nil
}