package core

import (
	"testing"
)

func TestParseTypeExpression(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
		check   func(*TypeDef) bool
	}{
		{
			name: "simple string",
			expr: "string",
			check: func(td *TypeDef) bool {
				return td.Kind == KindString
			},
		},
		{
			name: "string with length range",
			expr: "string(1..100)",
			check: func(td *TypeDef) bool {
				return td.Kind == KindString &&
					*td.MinLength == 1 &&
					*td.MaxLength == 100
			},
		},
		{
			name: "string with format",
			expr: "string(email)",
			check: func(td *TypeDef) bool {
				return td.Kind == KindString && td.Format == "email"
			},
		},
		{
			name: "integer with range",
			expr: "int(0..999)",
			check: func(td *TypeDef) bool {
				return td.Kind == KindInt &&
					*td.Min == 0 &&
					*td.Max == 999
			},
		},
		{
			name: "enum",
			expr: "enum(draft, published, archived)",
			check: func(td *TypeDef) bool {
				return td.Kind == KindEnum &&
					len(td.Enum) == 3 &&
					td.Enum[0] == "draft"
			},
		},
		{
			name: "reference",
			expr: "$User",
			check: func(td *TypeDef) bool {
				return td.Kind == KindRef && td.Ref == "$User"
			},
		},
		{
			name: "array of strings",
			expr: "string[]",
			check: func(td *TypeDef) bool {
				return td.Kind == KindArray &&
					td.Items != nil &&
					td.Items.Kind == KindString
			},
		},
		{
			name: "array of references",
			expr: "$User[]",
			check: func(td *TypeDef) bool {
				return td.Kind == KindArray &&
					td.Items != nil &&
					td.Items.Kind == KindRef &&
					td.Items.Ref == "$User"
			},
		},
		{
			name: "array with max constraint",
			expr: "string[](max:10)",
			check: func(td *TypeDef) bool {
				return td.Kind == KindArray &&
					td.Items.Kind == KindString &&
					*td.MaxItems == 10
			},
		},
		{
			name: "map type",
			expr: "map(string, any)",
			check: func(td *TypeDef) bool {
				return td.Kind == KindMap &&
					td.ValueType != nil &&
					td.ValueType.Kind == KindAny
			},
		},
		{
			name: "datetime",
			expr: "datetime",
			check: func(td *TypeDef) bool {
				return td.Kind == KindDatetime
			},
		},
		{
			name: "uuid",
			expr: "uuid",
			check: func(td *TypeDef) bool {
				return td.Kind == KindUUID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td, err := ParseTypeExpressionFull(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTypeExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !tt.check(td) {
				t.Errorf("ParseTypeExpression() check failed for %s", tt.expr)
			}
		})
	}
}

func TestParseTypes(t *testing.T) {
	parser := NewTypeParser()

	// Тест парсинга полного типа
	types := map[string]interface{}{
		"User": map[string]interface{}{
			"id":         "uuid",
			"name":       "string(1..100)",
			"email":      "string(email)",
			"age":        "int(0..120)",
			"roles":      "string[]",
			"created_at": "datetime",
		},
		"Post": map[string]interface{}{
			"id":        "uuid",
			"title":     "string(1..200)",
			"content":   "string(1..10000)",
			"author":    "$User",
			"tags":      "string[](max:10)",
			"status":    "enum(draft, published, archived)",
			"metadata":  "map(string, any)",
		},
		"Comment": map[string]interface{}{
			"id":      "uuid",
			"text":    "string(1..1000)",
			"author":  "$User",
			"post":    "$Post",
			"replies": "$Comment[]",
		},
	}

	registry, err := parser.ParseTypes(types)
	if err != nil {
		t.Fatalf("ParseTypes() error = %v", err)
	}

	// Проверяем, что типы созданы
	if len(registry.Types) != 3 {
		t.Errorf("Expected 3 types, got %d", len(registry.Types))
	}

	// Проверяем тип User
	userType, ok := registry.Types["User"]
	if !ok {
		t.Fatal("User type not found")
	}
	if userType.Name != "User" {
		t.Errorf("Expected User name, got %s", userType.Name)
	}
	if userType.Kind != KindObject {
		t.Errorf("Expected object kind, got %s", userType.Kind)
	}
	if len(userType.Properties) != 6 {
		t.Errorf("Expected 6 properties for User, got %d", len(userType.Properties))
	}

	// Проверяем поле email
	emailField := userType.Properties["email"]
	if emailField == nil {
		t.Fatal("email field not found")
	}
	if emailField.Kind != KindString {
		t.Errorf("Expected string kind for email, got %s", emailField.Kind)
	}
	if emailField.Format != "email" {
		t.Errorf("Expected email format, got %s", emailField.Format)
	}

	// Проверяем тип Post
	postType, ok := registry.Types["Post"]
	if !ok {
		t.Fatal("Post type not found")
	}

	// Проверяем поле author (ссылка)
	authorField := postType.Properties["author"]
	if authorField == nil {
		t.Fatal("author field not found")
	}
	if authorField.Kind != KindRef {
		t.Errorf("Expected ref kind for author, got %s", authorField.Kind)
	}
	if authorField.Ref != "$User" {
		t.Errorf("Expected $User ref, got %s", authorField.Ref)
	}

	// Проверяем поле status (enum)
	statusField := postType.Properties["status"]
	if statusField == nil {
		t.Fatal("status field not found")
	}
	if statusField.Kind != KindEnum {
		t.Errorf("Expected enum kind for status, got %s", statusField.Kind)
	}
	if len(statusField.Enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(statusField.Enum))
	}

	// Проверяем рекурсивный тип Comment
	commentType, ok := registry.Types["Comment"]
	if !ok {
		t.Fatal("Comment type not found")
	}

	// Проверяем поле replies (массив ссылок на себя)
	repliesField := commentType.Properties["replies"]
	if repliesField == nil {
		t.Fatal("replies field not found")
	}
	if repliesField.Kind != KindArray {
		t.Errorf("Expected array kind for replies, got %s", repliesField.Kind)
	}
	if repliesField.Items == nil || repliesField.Items.Kind != KindRef {
		t.Error("Expected ref items for replies array")
	}
	if repliesField.Items.Ref != "$Comment" {
		t.Errorf("Expected $Comment ref for items, got %s", repliesField.Items.Ref)
	}
}

func TestTypeRegistryResolve(t *testing.T) {
	parser := NewTypeParser()

	// Создаём основной модуль
	mainTypes := map[string]interface{}{
		"User": map[string]interface{}{
			"id":   "uuid",
			"name": "string",
		},
	}

	mainRegistry, err := parser.ParseTypes(mainTypes)
	if err != nil {
		t.Fatalf("ParseTypes() error = %v", err)
	}

	// Создаём импортированный модуль
	commonParser := NewTypeParser()
	commonTypes := map[string]interface{}{
		"Timestamp": map[string]interface{}{
			"created_at": "datetime",
			"updated_at": "datetime",
		},
	}

	commonRegistry, err := commonParser.ParseTypes(commonTypes)
	if err != nil {
		t.Fatalf("ParseTypes() error = %v", err)
	}

	// Добавляем импорт
	mainRegistry.Imports["common"] = commonRegistry

	// Тестируем разрешение локального типа
	userType, err := mainRegistry.Resolve("User")
	if err != nil {
		t.Errorf("Failed to resolve User: %v", err)
	}
	if userType.Name != "User" {
		t.Errorf("Expected User, got %s", userType.Name)
	}

	// Тестируем разрешение импортированного типа
	timestampType, err := mainRegistry.Resolve("common.Timestamp")
	if err != nil {
		t.Errorf("Failed to resolve common.Timestamp: %v", err)
	}
	if timestampType.Name != "Timestamp" {
		t.Errorf("Expected Timestamp, got %s", timestampType.Name)
	}

	// Тестируем ошибку при несуществующем типе
	_, err = mainRegistry.Resolve("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent type")
	}

	// Тестируем ошибку при несуществующем модуле
	_, err = mainRegistry.Resolve("unknown.Type")
	if err == nil {
		t.Error("Expected error for unknown module")
	}
}