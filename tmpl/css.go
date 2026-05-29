package tmpl

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/tomsiouan/templify/config"
)

// buildConfigCSS generates a <style> block from cfg, covering CSS variables,
// page layout, typography, heading sizes, and optional paged.js band styles.
func buildConfigCSS(cfg *config.Config) template.HTML {
	var sb strings.Builder
	sb.WriteString("<style>\n")

	writeCSSVariables(&sb, cfg)
	writePageRules(&sb, cfg)
	writeBodyRules(&sb, cfg)

	if cfg.HeadingNumbers.Enabled {
		writeHeadingNumberCounters(&sb)
	}

	writeHeadingSizes(&sb, cfg)

	if cfg.Header.Enabled && cfg.Header.Background != "" {
		writeHeaderBandCSS(&sb, cfg.Header.Background)
	}

	if cfg.Footer.Enabled && cfg.Footer.Background != "" {
		writeFooterBandCSS(&sb, cfg.Footer.Background)
	}

	sb.WriteString("</style>")

	return template.HTML(sb.String())
}

// writeCSSVariables writes the :root block mapping config colors to CSS custom properties.
func writeCSSVariables(sb *strings.Builder, cfg *config.Config) {
	fmt.Fprintf(sb, ":root {\n")
	fmt.Fprintf(sb, "    --blue: %s;\n", cfg.Colors.Primary)
	fmt.Fprintf(sb, "    --blue-light: %s;\n", cfg.Colors.PrimaryLight)
	fmt.Fprintf(sb, "    --bg-soft: %s;\n", cfg.Colors.Background)
	fmt.Fprintf(sb, "    --text: %s;\n", cfg.Colors.Text)
	fmt.Fprintf(sb, "    --text-muted: %s;\n", cfg.Colors.TextMuted)
	fmt.Fprintf(sb, "}\n\n")
}

// writePageRules writes @page rules for the default, cover, blank, and toc named pages.
func writePageRules(sb *strings.Builder, cfg *config.Config) {
	m := cfg.Page.Margins
	suppress := suppressAllMarginBoxes()

	fmt.Fprintf(sb, "@page {\n")
	fmt.Fprintf(sb, "    size: %s;\n", cfg.Page.Size)
	fmt.Fprintf(sb, "    margin: %s %s %s %s;\n", m.Top, m.Right, m.Bottom, m.Left)
	writeMarginBoxes(sb, cfg)
	fmt.Fprintf(sb, "}\n\n")

	fmt.Fprintf(sb, "@page cover { size: %s; margin: 0; %s }\n", cfg.Page.Size, suppress)
	fmt.Fprintf(sb, "@page blank { size: %s; margin: %s %s %s %s; %s }\n\n",
		cfg.Page.Size, m.Top, m.Right, m.Bottom, m.Left, suppress)
	fmt.Fprintf(sb, "@page toc { size: %s; margin: %s %s %s %s; %s }\n\n",
		cfg.Page.Size, m.Top, m.Right, m.Bottom, m.Left, suppress)
}

// writeBodyRules writes body typography and optional justify, paragraph-indent,
// and heading-indent rules.
func writeBodyRules(sb *strings.Builder, cfg *config.Config) {
	fmt.Fprintf(sb, "body {\n")
	fmt.Fprintf(sb, "    font-family: \"%s\", sans-serif;\n", cfg.Font.Family)
	fmt.Fprintf(sb, "    font-size: %s;\n", cfg.Font.Size)
	fmt.Fprintf(sb, "    line-height: %g;\n", cfg.Font.LineHeight)
	fmt.Fprintf(sb, "}\n\n")
	if cfg.Justify {
		fmt.Fprintf(sb, "p, li, dt, dd, blockquote p {\n")
		fmt.Fprintf(sb, "    text-align: justify;\n")
		fmt.Fprintf(sb, "    hyphens: auto;\n")
		fmt.Fprintf(sb, "}\n\n")
	}
	if cfg.ParagraphIndent != "" {
		fmt.Fprintf(sb, "p { text-indent: %s; }\n\n", cfg.ParagraphIndent)
	}
	if cfg.HeadingIndent != "" {
		fmt.Fprintf(sb, "h3 { padding-left: %s; }\n", cfg.HeadingIndent)
		fmt.Fprintf(sb, "h4 { padding-left: calc(2 * %s); }\n", cfg.HeadingIndent)
		fmt.Fprintf(sb, "h5 { padding-left: calc(3 * %s); }\n", cfg.HeadingIndent)
		fmt.Fprintf(sb, "h6 { padding-left: calc(4 * %s); }\n\n", cfg.HeadingIndent)
	}
}

// writeHeadingNumberCounters writes CSS counter rules that auto-number h2-h4
// headings and their corresponding TOC entries.
func writeHeadingNumberCounters(sb *strings.Builder) {
	sb.WriteString("body { counter-reset: h2c; }\n")
	sb.WriteString("h2:not(.no-number) { counter-reset: h3c; counter-increment: h2c; }\n")
	sb.WriteString("h2:not(.no-number)::before { content: counter(h2c) \". \"; }\n")
	sb.WriteString("h3:not(.no-number) { counter-reset: h4c; counter-increment: h3c; }\n")
	sb.WriteString("h3:not(.no-number)::before { content: counter(h2c) \".\" counter(h3c) \" \"; }\n")
	sb.WriteString("h4:not(.no-number) { counter-increment: h4c; }\n")
	sb.WriteString("h4:not(.no-number)::before { content: counter(h2c) \".\" counter(h3c) \".\" counter(h4c) \" \"; }\n")
	sb.WriteString(".toc-page { counter-reset: th2c; }\n")
	sb.WriteString(".toc-level-2:not(.no-number) { counter-reset: th3c; counter-increment: th2c; }\n")
	sb.WriteString(".toc-level-2:not(.no-number) .toc-entry-title::before { content: counter(th2c) \". \"; }\n")
	sb.WriteString(".toc-level-3:not(.no-number) { counter-reset: th4c; counter-increment: th3c; }\n")
	sb.WriteString(".toc-level-3:not(.no-number) .toc-entry-title::before { content: counter(th2c) \".\" counter(th3c) \" \"; }\n")
	sb.WriteString(".toc-level-4:not(.no-number) { counter-increment: th4c; }\n")
	sb.WriteString(".toc-level-4:not(.no-number) .toc-entry-title::before { content: counter(th2c) \".\" counter(th3c) \".\" counter(th4c) \" \"; }\n\n")
}

// writeHeadingSizes writes font-size rules for h1-h6, adding page-break-before
// for any heading level configured with PageBreakBefore.
func writeHeadingSizes(sb *strings.Builder, cfg *config.Config) {
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
		fmt.Fprintf(sb, "%s { font-size: %s;", h.tag, h.level.Size)
		if h.level.PageBreakBefore {
			sb.WriteString(" break-before: page; page-break-before: always;")
		}
		sb.WriteString(" }\n")
	}
}

// writeHeaderBandCSS writes paged.js rules that render a full-width colored
// band behind the header margin boxes using a pseudo-element.
func writeHeaderBandCSS(sb *strings.Builder, background string) {
	fmt.Fprintf(sb, ".pagedjs_margin-top { position: relative; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-top::before {\n")
	fmt.Fprintf(sb, "    content: \"\";\n")
	fmt.Fprintf(sb, "    position: absolute;\n")
	fmt.Fprintf(sb, "    top: 3mm; height: 8mm;\n")
	fmt.Fprintf(sb, "    left: calc(-1 * var(--pagedjs-margin-left));\n")
	fmt.Fprintf(sb, "    right: calc(-1 * var(--pagedjs-margin-right));\n")
	fmt.Fprintf(sb, "    background: %s;\n", background)
	fmt.Fprintf(sb, "    z-index: 0;\n")
	fmt.Fprintf(sb, "}\n")
	fmt.Fprintf(sb, ".pagedjs_margin-top .pagedjs_margin { background: transparent; position: relative; z-index: 1; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-top .pagedjs_margin-content { position: absolute; top: 3mm; height: 8mm; display: flex; align-items: center; left: 0; right: 0; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-top-left .pagedjs_margin-content { justify-content: flex-start; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-top-center .pagedjs_margin-content { justify-content: center; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-top-right .pagedjs_margin-content { justify-content: flex-end; }\n")
	fmt.Fprintf(sb, ".pagedjs_named_page .pagedjs_margin-top::before, .pagedjs_blank_page .pagedjs_margin-top::before { display: none; }\n\n")
}

// writeFooterBandCSS writes paged.js rules that render a full-width colored
// band behind the footer margin boxes using a pseudo-element.
func writeFooterBandCSS(sb *strings.Builder, background string) {
	fmt.Fprintf(sb, ".pagedjs_margin-bottom { position: relative; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-bottom::before {\n")
	fmt.Fprintf(sb, "    content: \"\";\n")
	fmt.Fprintf(sb, "    position: absolute;\n")
	fmt.Fprintf(sb, "    bottom: 3mm; height: 8mm;\n")
	fmt.Fprintf(sb, "    left: calc(-1 * var(--pagedjs-margin-left));\n")
	fmt.Fprintf(sb, "    right: calc(-1 * var(--pagedjs-margin-right));\n")
	fmt.Fprintf(sb, "    background: %s;\n", background)
	fmt.Fprintf(sb, "    z-index: 0;\n")
	fmt.Fprintf(sb, "}\n")
	fmt.Fprintf(sb, ".pagedjs_margin-bottom .pagedjs_margin { background: transparent; position: relative; z-index: 1; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-bottom .pagedjs_margin-content { position: absolute; bottom: 3mm; height: 8mm; display: flex; align-items: center; left: 0; right: 0; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-bottom-left .pagedjs_margin-content { justify-content: flex-start; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-bottom-center .pagedjs_margin-content { justify-content: center; }\n")
	fmt.Fprintf(sb, ".pagedjs_margin-bottom-right .pagedjs_margin-content { justify-content: flex-end; }\n")
	fmt.Fprintf(sb, ".pagedjs_named_page .pagedjs_margin-bottom::before, .pagedjs_blank_page .pagedjs_margin-bottom::before { display: none; }\n\n")
}

// writeMarginBoxes writes @top-* and @bottom-* margin box rules inside an @page block.
func writeMarginBoxes(sb *strings.Builder, cfg *config.Config) {
	hdr := cfg.Header
	ftr := cfg.Footer
	hStyle := marginBoxStyle(cfg.Font.Family)
	fStyle := marginBoxStyle(cfg.Font.Family)

	if hdr.Enabled {
		fmt.Fprintf(sb, "    @top-left    { content: %s; %s text-align: left; }\n", config.ContentCSS(hdr.Left), hStyle)
		fmt.Fprintf(sb, "    @top-center  { content: %s; %s text-align: center; }\n", config.ContentCSS(hdr.Center), hStyle)
		fmt.Fprintf(sb, "    @top-right   { content: %s; %s text-align: right; }\n", config.ContentCSS(hdr.Right), hStyle)
	} else {
		sb.WriteString("    @top-left { content: none; } @top-center { content: none; } @top-right { content: none; }\n")
	}

	if ftr.Enabled {
		fmt.Fprintf(sb, "    @bottom-left   { content: %s; %s text-align: left; }\n", config.ContentCSS(ftr.Left), fStyle)
		fmt.Fprintf(sb, "    @bottom-center { content: %s; %s text-align: center; }\n", config.ContentCSS(ftr.Center), fStyle)
		fmt.Fprintf(sb, "    @bottom-right  { content: %s; %s text-align: right; }\n", config.ContentCSS(ftr.Right), fStyle)
	} else {
		sb.WriteString("    @bottom-left { content: none; } @bottom-center { content: none; } @bottom-right { content: none; }\n")
	}
}

// marginBoxStyle returns inline CSS for margin box font and color.
func marginBoxStyle(fontFamily string) string {
	return fmt.Sprintf(`font-family: "%s", sans-serif; font-size: 8pt; color: var(--text-muted);`, fontFamily)
}

// suppressAllMarginBoxes returns a CSS fragment that hides all six margin boxes.
func suppressAllMarginBoxes() string {
	boxes := []string{
		"@top-left", "@top-center", "@top-right",
		"@bottom-left", "@bottom-center", "@bottom-right",
	}
	var parts []string
	for _, b := range boxes {
		parts = append(parts, b+" { content: none; }")
	}
	return strings.Join(parts, " ")
}
