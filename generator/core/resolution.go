package core

import (
	"fmt"
	"strings"
)

// ResolveSpec resolves all types and references in a spec
func ResolveSpec(spec *Spec) error {
	if spec == nil {
		return fmt.Errorf("spec is nil")
	}

	// Initialize resolution if needed
	if spec.Resolved.TypeRegistry == nil {
		spec.Resolved.TypeRegistry = &TypeRegistry{
			Types: make(map[string]*TypeDef),
		}
	}

	// Parse all types using TypeParser
	parser := NewTypeParser()
	registry, err := parser.ParseTypes(spec.Types)
	if err != nil {
		return fmt.Errorf("failed to parse types: %w", err)
	}
	spec.Resolved.TypeRegistry = registry

	// Resolve references in types
	for _, td := range spec.Resolved.TypeRegistry.Types {
		if err := resolveTypeRefs(td, spec.Resolved.TypeRegistry); err != nil {
			return err
		}
	}

	// Resolve assistant input/output types
	for name, assistant := range spec.Assistants {
		// Resolve input type
		if assistant.InputType != "" {
			inputType, err := resolveTypeByName(assistant.InputType, spec.Resolved.TypeRegistry)
			if err != nil {
				return fmt.Errorf("assistant %s: failed to resolve input type %s: %w",
					name, assistant.InputType, err)
			}
			assistant.Resolved.InputType = inputType
		}

		// Resolve output type (по умолчанию string если не указан)
		outputTypeName := assistant.OutputType
		if outputTypeName == "" {
			outputTypeName = "string"
		}
		outputType, err := resolveTypeByName(outputTypeName, spec.Resolved.TypeRegistry)
		if err != nil {
			return fmt.Errorf("assistant %s: failed to resolve output type %s: %w",
				name, outputTypeName, err)
		}
		assistant.Resolved.OutputType = outputType

		// Validate dialog configuration
		if assistant.Dialog != nil && assistant.Thread == nil {
			return fmt.Errorf("assistant %s: dialog mode requires thread configuration (add 'thread' field)", name)
		}

		// Update the assistant in the map with resolved types
		spec.Assistants[name] = assistant
	}

	return nil
}

// resolveTypeByName finds a type by name or creates an inline type
func resolveTypeByName(typeName string, registry *TypeRegistry) (*TypeDef, error) {
	// Check if it's a reference to existing type
	if td, ok := registry.Types[typeName]; ok {
		return td, nil
	}

	// Try to parse as inline type expression
	td, err := ParseTypeExpression(typeName)
	if err != nil {
		return nil, fmt.Errorf("type %s not found and cannot parse as inline type: %w", typeName, err)
	}

	return td, nil
}

// resolveTypeRefs resolves all references within a type
func resolveTypeRefs(td *TypeDef, registry *TypeRegistry) error {
	if td == nil {
		return nil
	}

	switch td.Kind {
	case KindRef:
		// Validate that reference exists
		refName := strings.TrimPrefix(td.Ref, "$")
		if refName != "" {
			if _, ok := registry.Types[refName]; !ok {
				return fmt.Errorf("reference to undefined type: %s", td.Ref)
			}
		}

	case KindArray:
		if td.Items != nil {
			if err := resolveTypeRefs(td.Items, registry); err != nil {
				return err
			}
		}

	case KindMap:
		if td.ValueType != nil {
			if err := resolveTypeRefs(td.ValueType, registry); err != nil {
				return err
			}
		}

	case KindObject:
		for fieldName, fieldType := range td.Properties {
			if err := resolveTypeRefs(fieldType, registry); err != nil {
				return fmt.Errorf("field %s: %w", fieldName, err)
			}
		}
	}

	return nil
}