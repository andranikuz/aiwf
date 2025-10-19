package core

import (
	"fmt"
	"strings"
)

// TypeDef описывает тип в упрощённом формате AIWF
type TypeDef struct {
	// Имя типа (для именованных типов в реестре)
	Name string

	// ID типа для уникальной идентификации (например, aiwf://module/TypeName)
	ID string

	// Базовый тип или ссылка
	Kind TypeKind

	// Для примитивов
	Format    string   // email, url, phone, uuid, datetime, date
	Min       *float64 // для number/int
	Max       *float64
	MinLength *int     // для string
	MaxLength *int
	Pattern   string   // regex для string
	Enum      []string // для перечислений

	// Для объектов
	Properties map[string]*TypeDef

	// Для массивов
	Items    *TypeDef
	MinItems *int
	MaxItems *int

	// Ссылка на другой тип
	Ref string // $TypeName или $module.TypeName

	// Для map типов
	ValueType *TypeDef // map(string, ValueType)

	// Метаданные
	Description string
	Required    bool // все поля обязательные по умолчанию
	Optional    bool // поле помечено как опциональное (с ? суффиксом)
}

// TypeKind определяет вид типа
type TypeKind string

const (
	KindString   TypeKind = "string"
	KindInt      TypeKind = "int"
	KindNumber   TypeKind = "number"
	KindBool     TypeKind = "bool"
	KindDatetime TypeKind = "datetime"
	KindDate     TypeKind = "date"
	KindUUID     TypeKind = "uuid"
	KindObject   TypeKind = "object"
	KindArray    TypeKind = "array"
	KindMap      TypeKind = "map"
	KindEnum     TypeKind = "enum"
	KindRef      TypeKind = "ref"
	KindAny      TypeKind = "any"
)

// TypeRegistry хранит все определённые типы
type TypeRegistry struct {
	Types   map[string]*TypeDef
	Imports map[string]*TypeRegistry // импортированные модули
}

// Resolve находит тип по имени, включая импортированные
func (r *TypeRegistry) Resolve(ref string) (*TypeDef, error) {
	// Убираем префикс $
	ref = strings.TrimPrefix(ref, "$")

	// Проверяем, есть ли namespace (module.Type)
	if idx := strings.Index(ref, "."); idx > 0 {
		module := ref[:idx]
		typeName := ref[idx+1:]

		if imported, ok := r.Imports[module]; ok {
			if typ, ok := imported.Types[typeName]; ok {
				return typ, nil
			}
			return nil, fmt.Errorf("type %s not found in module %s", typeName, module)
		}
		return nil, fmt.Errorf("module %s not imported", module)
	}

	// Локальный тип
	if typ, ok := r.Types[ref]; ok {
		return typ, nil
	}

	return nil, fmt.Errorf("type %s not found", ref)
}

// ParseTypeExpression разбирает выражение типа вида "string(1..100)" или "$User[]"
func ParseTypeExpression(expr string) (*TypeDef, error) {
	expr = strings.TrimSpace(expr)

	// Проверяем на массив
	if strings.HasSuffix(expr, "[]") {
		baseExpr := strings.TrimSuffix(expr, "[]")
		itemType, err := ParseTypeExpression(baseExpr)
		if err != nil {
			return nil, err
		}
		return &TypeDef{
			Kind:  KindArray,
			Items: itemType,
		}, nil
	}

	// Проверяем на массив с ограничениями: Type[](max:10)
	if idx := strings.Index(expr, "[]("); idx > 0 {
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

		// Парсим ограничения
		if strings.HasPrefix(constraints, "max:") {
			// TODO: реализовать парсинг max
		}

		return arrayType, nil
	}

	// Проверяем на ссылку
	if strings.HasPrefix(expr, "$") {
		return &TypeDef{
			Kind: KindRef,
			Ref:  expr,
		}, nil
	}

	// Проверяем на enum
	if strings.HasPrefix(expr, "enum(") && strings.HasSuffix(expr, ")") {
		values := expr[5 : len(expr)-1]
		items := strings.Split(values, ",")
		for i := range items {
			items[i] = strings.TrimSpace(items[i])
		}
		return &TypeDef{
			Kind: KindEnum,
			Enum: items,
		}, nil
	}

	// Проверяем на map
	if strings.HasPrefix(expr, "map(") && strings.HasSuffix(expr, ")") {
		// map(string, any) -> ValueType = any
		content := expr[4 : len(expr)-1]
		parts := strings.SplitN(content, ",", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) != "string" {
			return nil, fmt.Errorf("map must have string keys: %s", expr)
		}

		valueType, err := ParseTypeExpression(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}

		return &TypeDef{
			Kind:      KindMap,
			ValueType: valueType,
		}, nil
	}

	// Проверяем на тип с ограничениями: string(1..100) или string(email)
	if idx := strings.Index(expr, "("); idx > 0 && strings.HasSuffix(expr, ")") {
		baseType := expr[:idx]
		constraints := expr[idx+1 : len(expr)-1]

		td := &TypeDef{}

		switch baseType {
		case "string":
			td.Kind = KindString
			// Парсим ограничения: может быть формат или диапазон
			if strings.Contains(constraints, "..") {
				// Диапазон: 1..100
				// TODO: реализовать парсинг диапазона
			} else {
				// Формат: email, url, phone
				td.Format = constraints
			}
		case "int":
			td.Kind = KindInt
			// TODO: парсинг диапазона для int
		case "number":
			td.Kind = KindNumber
			// TODO: парсинг диапазона для number
		default:
			return nil, fmt.Errorf("unknown type with constraints: %s", baseType)
		}

		return td, nil
	}

	// Простые типы без ограничений
	switch expr {
	case "string":
		return &TypeDef{Kind: KindString}, nil
	case "int":
		return &TypeDef{Kind: KindInt}, nil
	case "number":
		return &TypeDef{Kind: KindNumber}, nil
	case "bool":
		return &TypeDef{Kind: KindBool}, nil
	case "datetime":
		return &TypeDef{Kind: KindDatetime}, nil
	case "date":
		return &TypeDef{Kind: KindDate}, nil
	case "uuid":
		return &TypeDef{Kind: KindUUID}, nil
	case "any":
		return &TypeDef{Kind: KindAny}, nil
	default:
		// Treat unknown types as references
		return &TypeDef{
			Kind: KindRef,
			Ref:  expr,
		}, nil
	}
}