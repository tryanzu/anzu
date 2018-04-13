package templates

import (
	"bytes"
	"html/template"
	"strings"
)

var GlobalFuncs = template.FuncMap{
	"trust": func(html string) template.HTML {
		return template.HTML(html)
	},
	"nl2br": func(html template.HTML) template.HTML {
		return template.HTML(strings.Replace(string(html), "\n", "<br />", -1))
	},
}

var Templates = template.Must(template.New("").Funcs(GlobalFuncs).ParseGlob("./static/templates/**/*.html"))

func Execute(name string, data interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tpl, err := template.New(name).Funcs(GlobalFuncs).ParseFiles(name)
	if err != nil {
		return nil, err
	}
	err = tpl.Execute(buf, data)
	return buf, err
}

func ExecuteTemplate(name string, data interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	err := Templates.ExecuteTemplate(buf, name, data)
	return buf, err
}
