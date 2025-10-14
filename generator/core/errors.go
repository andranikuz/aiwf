package core

// MultiError аккумулирует несколько ValidationError.
type MultiError struct {
	Errors   []*ValidationError
	Warnings []*ValidationWarning
}

func (m *MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}
	return "multiple validation errors"
}

func (m *MultiError) Append(err error) {
	if err == nil {
		return
	}
	if ve, ok := err.(*ValidationError); ok {
		m.Errors = append(m.Errors, ve)
		return
	}
	if multi, ok := err.(*MultiError); ok {
		m.Errors = append(m.Errors, multi.Errors...)
		return
	}
	m.Errors = append(m.Errors, &ValidationError{Msg: err.Error()})
}

func (m *MultiError) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *MultiError) AppendWarning(w *ValidationWarning) {
	if w == nil {
		return
	}
	m.Warnings = append(m.Warnings, w)
}

func (m *MultiError) HasWarnings() bool {
	return len(m.Warnings) > 0
}
