package server

import (
	"html/template"

	"gopkg.in/yaml.v2"
)

func createTemplate(name string, tpl string) (*template.Template, error) {
	return template.New(name).Funcs(template.FuncMap{
		"toYaml": toYaml,
	}).Parse(tpl)
}

func toYaml(v interface{}) (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
