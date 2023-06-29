package utils

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"
)

func HCLTemplate(data string, params map[string]any) string {
	var out bytes.Buffer
	tmpl := template.Must(template.New("hcl").Parse(data))
	err := tmpl.Execute(&out, params)
	if err != nil {
		panic(err)
	}
	return out.String()
}

func HCLTemplateFiles(files ...string) func(params any) string {
	tmpl := template.Must(template.ParseFiles(files...))

	return func(params any) string {
		var out bytes.Buffer
		err := tmpl.Execute(&out, params)
		if err != nil {
			panic(err)
		}
		return out.String()
	}
}

func HCLTemplateGlob(glob string) func(params any) string {
	filenames, err := filepath.Glob(glob)
	if err != nil {
		panic(err)
	}
	if len(filenames) == 0 {
		panic(fmt.Sprintf("template: pattern matches no files: %#q", glob))
	}

	return HCLTemplateFiles(filenames...)
}
