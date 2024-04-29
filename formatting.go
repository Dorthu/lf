package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"text/template"

	"github.com/go-logfmt/logfmt"
)

// reForm is the regex used to replace format arguments with templated placeholders.
// For example, `[.time] .msg` would become `[{{.time}}] {{.msg}}`
var reForm = regexp.MustCompile(`(^|[^a-zA-Z0-9_])(\.[a-zA-Z0-9_]+)`)

// readFormat reads the incoming format string and returns a template
// that outputs in that format
func readFormat(format string) *template.Template {
	if format == "" {
		return nil
	}

	format = reForm.ReplaceAllString(format, "$1{{$2}}")
	template, err := template.New("format").Option("missingkey=zero").Parse(format)

	if err != nil {
		fmt.Printf("Invalid format stirng %s", format)
		panic(err)
	}

	return template
}

// formatLine accepts a format template and a parsed line and outputs the result
// to os.Stdout
func formatLine(data map[string]string, format *template.Template) {
	var buf bytes.Buffer
	if err := format.Execute(&buf, data); err == nil {
		fmt.Println(buf.String())
	}
}

// dump writes the full set of key/value pairs back out as logfmt
func dump(data map[string]string) {
	enc := logfmt.NewEncoder(os.Stdout)
	for k, v := range data {
		enc.EncodeKeyval(k, v)
	}
	enc.EndRecord()
}
