package tmpl

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomsiouan/templify/pkg/config"
	"github.com/tomsiouan/templify/pkg/parser"
)

//go:embed builtin/*.html
var builtinFS embed.FS

func Render(name string, doc *parser.Document, cfg *config.Config) (string, error) {
	content, err := loadTemplate(name)
	if err != nil {
		return "", err
	}

	t, err := template.New("doc").Parse(content)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := struct {
		*parser.Document
		Config    *config.Config
		ConfigCSS template.HTML
	}{doc, cfg, buildConfigCSS(cfg)}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
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

func buildConfigCSS(cfg *config.Config) template.HTML {
	var sb strings.Builder

	sb.WriteString("<style>\n")

	fmt.Fprintf(&sb, ":root {\n")
	fmt.Fprintf(&sb, "    --blue: %s;\n", cfg.Colors.Primary)
	fmt.Fprintf(&sb, "    --blue-light: %s;\n", cfg.Colors.PrimaryLight)
	fmt.Fprintf(&sb, "    --bg-soft: %s;\n", cfg.Colors.Background)
	fmt.Fprintf(&sb, "    --text: %s;\n", cfg.Colors.Text)
	fmt.Fprintf(&sb, "    --text-muted: %s;\n", cfg.Colors.TextMuted)
	fmt.Fprintf(&sb, "}\n\n")

	fmt.Fprintf(&sb, "@page {\n")
	fmt.Fprintf(&sb, "    size: %s;\n", cfg.Page.Size)
	fmt.Fprintf(&sb, "    margin: %s %s %s %s;\n",
		cfg.Page.Margins.Top, cfg.Page.Margins.Right,
		cfg.Page.Margins.Bottom, cfg.Page.Margins.Left)
	if cfg.PageNumbers.Enabled {
		fmt.Fprintf(&sb, "    @bottom-right {\n")
		fmt.Fprintf(&sb, "        content: %s;\n", config.PageNumberCSS(cfg.PageNumbers.Format))
		fmt.Fprintf(&sb, "        font-family: \"%s\", sans-serif;\n", cfg.Font.Family)
		fmt.Fprintf(&sb, "        font-size: 8.5pt;\n")
		fmt.Fprintf(&sb, "        color: var(--blue-light);\n")
		fmt.Fprintf(&sb, "    }\n")
	}
	fmt.Fprintf(&sb, "}\n\n")

	fmt.Fprintf(&sb, "@page cover { size: %s; margin: 0; @bottom-right { content: none; } }\n", cfg.Page.Size)
	fmt.Fprintf(&sb, "@page blank { size: %s; margin: %s %s %s %s; @bottom-right { content: none; } }\n\n",
		cfg.Page.Size,
		cfg.Page.Margins.Top, cfg.Page.Margins.Right,
		cfg.Page.Margins.Bottom, cfg.Page.Margins.Left)

	fmt.Fprintf(&sb, "body {\n")
	fmt.Fprintf(&sb, "    font-family: \"%s\", sans-serif;\n", cfg.Font.Family)
	fmt.Fprintf(&sb, "    font-size: %s;\n", cfg.Font.Size)
	fmt.Fprintf(&sb, "    line-height: %g;\n", cfg.Font.LineHeight)
	fmt.Fprintf(&sb, "}\n\n")
	if cfg.Justify {
		fmt.Fprintf(&sb, "p, li, dt, dd, blockquote p {\n")
		fmt.Fprintf(&sb, "    text-align: justify;\n")
		fmt.Fprintf(&sb, "    hyphens: auto;\n")
		fmt.Fprintf(&sb, "}\n\n")
	}

	type hEntry struct {
		tag   string
		level config.HeadingLevel
	}
	for _, h := range []hEntry{
		{"h1", cfg.Headings.H1},
		{"h2", cfg.Headings.H2},
		{"h3", cfg.Headings.H3},
		{"h4", cfg.Headings.H4},
		{"h5", cfg.Headings.H5},
		{"h6", cfg.Headings.H6},
	} {
		fmt.Fprintf(&sb, "%s { font-size: %s;", h.tag, h.level.Size)
		if h.level.PageBreakBefore {
			sb.WriteString(" break-before: page; page-break-before: always;")
		}
		sb.WriteString(" }\n")
	}

	sb.WriteString("</style>")
	return template.HTML(sb.String())
}
