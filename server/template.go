package server

import (
	"html/template"

	"gopkg.in/yaml.v2"
)

func createTemplate(name string, tpl string) (*template.Template, error) {
	return template.New(name).Funcs(template.FuncMap{
		"toYaml":              toYaml,
		"htmlUnescape":        htmlUnescape,
		"toHTMLUnescapedYaml": toHTMLUnescapedYaml,
	}).Parse(tpl)
}

func toYaml(v interface{}) string {
	b, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func htmlUnescape(s string) template.HTML {
	return template.HTML(s)
}

func toHTMLUnescapedYaml(v interface{}) template.HTML {
	return htmlUnescape(toYaml(v))
}
