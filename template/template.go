package template

import (
	_ "embed"
	"html/template"
)

//go:embed template.tmpl
var templateTmpl string

var AutoIndexTemplate = template.Must(template.New("autoIndexTemplate").Parse(templateTmpl))
