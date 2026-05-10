package tmpl

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"

	"github.com/tomsiouan/templify/pkg/config"
	"github.com/tomsiouan/templify/pkg/document"
)

var styleTagRe = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)

// Render executes mainContent (HTML template) against doc and cfg.
// coverContent is the raw cover template (may be empty).
func Render(mainContent, coverContent string, doc *document.Document, cfg *config.Config) (string, error) {
	coverHTML, coverCSS, err := renderCover(coverContent, doc, cfg)
	if err != nil {
		return "", err
	}

	t, err := template.New("doc").Funcs(templateFuncs()).Parse(mainContent)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := struct {
		*document.Document
		Config    *config.Config
		ConfigCSS template.HTML
		CoverHTML template.HTML
		CoverCSS  template.HTML
	}{doc, cfg, buildConfigCSS(cfg), coverHTML, coverCSS}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
