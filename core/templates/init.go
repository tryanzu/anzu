package templates

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/tryanzu/core/core/config"
)

var GlobalFuncs = template.FuncMap{
	"trust": func(html string) template.HTML {
		return template.HTML(html)
	},
	"nl2br": func(html template.HTML) template.HTML {
		return template.HTML(strings.Replace(string(html), "\n", "<br />", -1))
	},
}

var Templates = template.Must(template.New("").Funcs(GlobalFuncs).ParseGlob("../../static/templates/**/*.html"))

func Execute(name string, data interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tpl, err := template.New(name).Funcs(GlobalFuncs).ParseFiles(name)
	if err != nil {
		return nil, err
	}
	err = tpl.Execute(buf, data)
	return buf, err
}

func ExecuteTemplate(name string, data interface{}) (buf *bytes.Buffer, err error) {
	buf = new(bytes.Buffer)
	switch v := data.(type) {
	case map[string]interface{}:
		v["config"] = config.C.Copy()
		err = Templates.ExecuteTemplate(buf, name, v)
	default:
		err = Templates.ExecuteTemplate(buf, name, v)
	}
	return
}
