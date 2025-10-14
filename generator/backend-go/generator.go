package backendgo

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"text/template"
	"unicode"

	"github.com/andranikuz/aiwf/generator/core"
)

// Options описывает параметры генерации Go SDK.
type Options struct {
	Package   string
	OutputDir string
}

// Generate создает Go SDK на основе IR.
func Generate(ir *core.IR, opts Options) (map[string][]byte, error) {
	if ir == nil {
		return nil, fmt.Errorf("backend-go: ir is nil")
	}
	if opts.Package == "" {
		opts.Package = "aiwfgen"
	}

	ctx := buildContext(ir, opts.Package)

	outputs := map[string]string{
		"templates/service.go.tmpl":   filepath.Join(opts.OutputDir, "service.go"),
		"templates/agents.go.tmpl":    filepath.Join(opts.OutputDir, "agents.go"),
		"templates/workflows.go.tmpl": filepath.Join(opts.OutputDir, "workflows.go"),
		"templates/contracts.go.tmpl": filepath.Join(opts.OutputDir, "contracts.go"),
	}

	files := make(map[string][]byte, len(outputs))

	for tmplPath, outPath := range outputs {
		data, err := renderTemplate(tmplPath, ctx)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", tmplPath, err)
		}
		files[outPath] = data
	}

	return files, nil
}

type assistantCtx struct {
	Name                string
	MethodName          string
	InterfaceName       string
	StructName          string
	Model               string
	SystemPrompt        string
	InputSchemaRef      string
	OutputSchemaRef     string
	InputContract       contractType
	OutputContract      contractType
	OutputSchemaVar     string
	OutputSchemaLiteral string
	HasOutputSchema     bool
}

type workflowCtx struct {
	Name       string
	MethodName string
	RunnerName string
	StructName string
	Steps      []workflowStepCtx
	InputType  string
	OutputType string
}

type workflowStepCtx struct {
	Name           string
	Method         string
	InputType      string
	OutputType     string
	ResultVar      string
	TraceVar       string
	Needs          []string
	IsLast         bool
	InputSource    string
	AssignToResult bool
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func buildContext(ir *core.IR, pkg string) struct {
	Package    string
	Assistants []assistantCtx
	Workflows  []workflowCtx
} {
	ctx := struct {
		Package    string
		Assistants []assistantCtx
		Workflows  []workflowCtx
	}{
		Package: pkg,
	}

	assistantMap := make(map[string]assistantCtx)
	for _, name := range sortedKeys(ir.Assistants) {
		pascal := pascalCase(name)
		assistant := ir.Assistants[name]
		schemaLiteral := loadSchemaLiteral(assistant.OutputSchemaPath)
		schemaVar := unexport(pascal) + "OutputSchemaJSON"
		aCtx := assistantCtx{
			Name:                name,
			MethodName:          pascal,
			InterfaceName:       pascal + "Agent",
			StructName:          unexport(pascal) + "Agent",
			Model:               assistant.Model,
			SystemPrompt:        assistant.SystemPrompt,
			InputSchemaRef:      assistant.InputSchemaRef,
			OutputSchemaRef:     assistant.OutputSchemaRef,
			InputContract:       loadContract(assistant.InputSchemaPath, pascal+"Input"),
			OutputContract:      loadContract(assistant.OutputSchemaPath, pascal+"Output"),
			OutputSchemaVar:     schemaVar,
			OutputSchemaLiteral: schemaLiteral,
			HasOutputSchema:     schemaLiteral != "",
		}
		ctx.Assistants = append(ctx.Assistants, aCtx)
		assistantMap[name] = aCtx
	}

	for _, name := range sortedKeys(ir.Workflows) {
		pascal := pascalCase(name)
		wf := ir.Workflows[name]
		steps := make([]workflowStepCtx, 0, len(wf.Steps))
		wfInputType := "map[string]any"
		wfOutputType := "map[string]any"
		for idx, step := range wf.Steps {
			method := ""
			inputType := "map[string]any"
			outputType := "map[string]any"
			if ac, ok := assistantMap[step.Assistant]; ok {
				method = ac.MethodName
				inputType = ac.MethodName + "Input"
				outputType = ac.MethodName + "Output"
			}
			varName := unexport(pascalCase(step.Name)) + "Result"
			traceVar := varName + "Trace"
			src := "\"input\""
			if len(step.Needs) > 0 {
				src = fmt.Sprintf("\"%s\"", step.Needs[0])
			}

			if idx == 0 {
				wfInputType = inputType
			}

			isLast := idx == len(wf.Steps)-1
			if isLast {
				wfOutputType = outputType
			}

			assignResult := isLast && method != ""

			steps = append(steps, workflowStepCtx{
				Name:           step.Name,
				Method:         method,
				InputType:      inputType,
				OutputType:     outputType,
				ResultVar:      varName,
				TraceVar:       traceVar,
				Needs:          step.Needs,
				IsLast:         isLast,
				InputSource:    src,
				AssignToResult: assignResult,
			})
		}
		ctx.Workflows = append(ctx.Workflows, workflowCtx{
			Name:       name,
			MethodName: pascal,
			RunnerName: pascal + "Workflow",
			StructName: unexport(pascal) + "Workflow",
			InputType:  wfInputType,
			Steps:      steps,
			OutputType: wfOutputType,
		})
	}

	return ctx
}

func renderTemplate(name string, ctx any) ([]byte, error) {
	tmpl, err := template.ParseFS(templatesFS, name)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func pascalCase(input string) string {
	if input == "" {
		return ""
	}
	runes := []rune(input)
	out := make([]rune, 0, len(runes))
	upperNext := true
	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upperNext = true
			continue
		}
		if upperNext {
			r = unicode.ToUpper(r)
			upperNext = false
		} else {
			prev := runes[i-1]
			if unicode.IsUpper(r) && unicode.IsLower(prev) {
				// keep camelCase boundary uppercase
			} else {
				r = unicode.ToLower(r)
			}
		}
		out = append(out, r)
	}
	if len(out) == 0 {
		return ""
	}
	out[0] = unicode.ToUpper(out[0])
	return string(out)
}

func unexport(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
