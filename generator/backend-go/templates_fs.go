package backendgo

import "embed"

//go:embed templates/*.go.tmpl
var templatesFS embed.FS
