package validate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/andranikuz/aiwf/generator/core"
)

// formatError преобразует ошибку загрузки/IR в человекочитаемый вид.
func FormatError(err error) string {
    var ve *core.ValidationError
    if errors.As(err, &ve) {
        return fmt.Sprintf("✗ %s — %s", ve.Field, ve.Msg)
    }

	var me *core.MultiError
	if errors.As(err, &me) {
        lines := make([]string, 0, len(me.Errors)+len(me.Warnings))
        for _, item := range me.Errors {
            lines = append(lines, FormatError(item))
        }
        for _, warn := range me.Warnings {
            lines = append(lines, formatWarning(warn))
        }
        return strings.Join(lines, "\n")
	}

	if strings.Contains(err.Error(), "schema") {
		return fmt.Sprintf("✗ schema: %s", err.Error())
	}
	return fmt.Sprintf("✗ %s", err.Error())
}

func formatWarning(w *core.ValidationWarning) string {
	return fmt.Sprintf("⚠ %s — %s", w.Field, w.Msg)
}
