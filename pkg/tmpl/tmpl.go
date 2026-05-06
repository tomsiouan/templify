package tmpl

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/tomsiouan/templify/pkg/parser"
)

//go:embed builtin/*.html
var builtinFS embed.FS

func Render(name string, doc *parser.Document) (string, error) {
	content, err := loadTemplate(name)
	if err != nil {
		return "", err
	}

	t, err := template.New("doc").Parse(content)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, doc); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

func loadTemplate(name string) (string, error) {
	if filepath.Ext(name) == ".html" {
		data, err := os.ReadFile(name)
		if err != nil {
			return "", fmt.Errorf("read template %q: %w", name, err)
		}
		return string(data), nil
	}

	data, err := builtinFS.ReadFile("builtin/" + name + ".html")
	if err != nil {
		return "", fmt.Errorf("built-in template %q not found", name)
	}
	return string(data), nil
}
