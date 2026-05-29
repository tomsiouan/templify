package tmpl

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/tomsiouan/templify/config"
	"github.com/tomsiouan/templify/document"
)

// renderCover executes the cover template and returns the body HTML and any
// extracted <style> blocks separately so they can be injected into <head>.
func renderCover(coverContent string, doc *document.Document, cfg *config.Config) (html template.HTML, css template.HTML, err error) {
	if !cfg.Cover.Enabled || coverContent == "" {
		return "", "", nil
	}

	t, err := template.New("cover").Funcs(templateFuncs()).Parse(coverContent)
	if err != nil {
		return "", "", fmt.Errorf("parse cover template: %w", err)
	}

	data := struct {
		*document.Document
		Config *config.Config
	}{doc, cfg}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("execute cover template: %w", err)
	}

	rendered := buf.String()
	styles := styleTagRe.FindAllString(rendered, -1)
	body := styleTagRe.ReplaceAllString(rendered, "")

	return template.HTML(body), template.HTML(strings.Join(styles, "\n")), nil
}
