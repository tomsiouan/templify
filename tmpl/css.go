package tmpl

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/tomsiouan/templify/config"
	"github.com/tomsiouan/templify/highlight"
)

// buildConfigCSS generates a <style> block from cfg, covering CSS variables,
// page layout, typography, heading sizes, and optional paged.js band styles.
func buildConfigCSS(cfg *config.Config) template.HTML {
	var sb strings.Builder
	sb.WriteString("<style>\n")

	// @import has to come before any rule to be honoured at all.
	writeCodeFontImport(&sb, cfg)

	// Self-hosted faces, inlined as data URIs before anything uses them.
	if faces := cfg.FontFacesCSS(); faces != "" {
		sb.WriteString(faces)
	}

	writeCSSVariables(&sb, cfg)
	writePageRules(&sb, cfg)
	writeBodyRules(&sb, cfg)
	writeCodeRules(&sb, cfg)

	if cfg.HeadingNumbers.Enabled {
		writeHeadingNumberCounters(&sb)
	}

	writeHeadingSizes(&sb, cfg)
	writeHeadingKeepTogetherRules(&sb)

	if cfg.Header.Enabled && cfg.Header.Background != "" {
		writeHeaderBandCSS(&sb, cfg.Header.Background)
	}

	if cfg.Footer.Enabled && cfg.Footer.Background != "" {
		writeFooterBandCSS(&sb, cfg.Footer.Background)
	}

	sb.WriteString("</style>")

	return template.HTML(sb.String())
}

// roundPx wraps a font-size value in round(value, 1px) so it always resolves
// to a whole device pixel.
//
// Chromium prints any @font-face-loaded text (self-hosted or linked, static or
// variable) as one positioned run per glyph rather than one run per word. When
// the font-size computes to a fractional pixel value, consecutive glyph
// advances end up inconsistent by a sub-pixel amount, and some PDF readers —
// confirmed on Apple's PDFKit, which backs Preview and Quick Look — misorder
// or duplicate runs around that inconsistency. It shows up most visibly around
// the underscore glyph in monospace code (`disable_mlock` copies out as
// "disable" / "_" / "mlock" on separate lines). A whole-pixel size avoids the
// inconsistency entirely; round() is a no-op when the value is already whole,
// so this only ever helps.
func roundPx(value string) string {
	return fmt.Sprintf("round(%s, 1px)", value)
}

// writeCSSVariables writes the :root block mapping config colors to CSS custom properties.
func writeCSSVariables(sb *strings.Builder, cfg *config.Config) {
	fmt.Fprintf(sb, ":root {\n")
	fmt.Fprintf(sb, "    --blue: %s;\n", cfg.Colors.Primary)
	fmt.Fprintf(sb, "    --blue-light: %s;\n", cfg.Colors.PrimaryLight)
	fmt.Fprintf(sb, "    --bg-soft: %s;\n", cfg.Colors.Background)
	fmt.Fprintf(sb, "    --text: %s;\n", cfg.Colors.Text)
	fmt.Fprintf(sb, "    --text-muted: %s;\n", cfg.Colors.TextMuted)
	background, foreground := highlight.Palette(cfg.Code)
	fmt.Fprintf(sb, "    --code-bg: %s;\n", background)
	fmt.Fprintf(sb, "    --code-fg: %s;\n", foreground)
	fmt.Fprintf(sb, "    --code-font: %s;\n", codeFontStack(cfg.Code.FontFamily))
	fmt.Fprintf(sb, "    --code-font-size: %s;\n", roundPx(cfg.Code.FontSize))
	fmt.Fprintf(sb, "    --code-line-height: %g;\n", cfg.Code.LineHeight)
	fmt.Fprintf(sb, "}\n\n")
}

// codeFontStack builds the code font-family value: the configured family first,
// then the monospace fallbacks so blocks never drop to a proportional font.
func codeFontStack(family string) string {
	family = strings.TrimSpace(family)
	if family == "" {
		return highlight.MonoFallbacks
	}
	return fmt.Sprintf("%q, %s", family, highlight.MonoFallbacks)
}

// writeCodeFontImport imports the webfont used by code blocks, when configured.
// Self-hosted faces win: importing the stylesheet as well would hand Chromium a
// second copy of the family and put the broken text layer back.
func writeCodeFontImport(sb *strings.Builder, cfg *config.Config) {
	if len(cfg.Code.Faces) > 0 {
		return
	}
	if url := strings.TrimSpace(cfg.Code.FontURL); url != "" {
		fmt.Fprintf(sb, "@import url(%q);\n\n", url)
	}
}

// writeCodeRules writes the code block stylesheet: the theme's token colors
// followed by the panel itself. Blocks are a self-contained filled panel rather
// than the left-bar treatment blockquotes use, so the two never look alike.
//
// The panel selectors are prefixed with `html` on purpose. Bundles are free to
// restyle anything else, but the token colors and the panel background are two
// halves of one theme: a bundle that repaints only the background (as every
// bundle copied from the pre-`code:` templates does) would leave light-on-light
// or dark-on-dark code. The prefix outranks a bare `pre` rule wherever it sits
// in the cascade, so the theme stays internally consistent; a bundle that
// really wants the panel can still win with `html pre` of its own.
func writeCodeRules(sb *strings.Builder, cfg *config.Config) {
	if css := highlight.CSS(cfg.Code); css != "" {
		sb.WriteString(css)
		sb.WriteString("\n")
	}

	// pre carries the panel, highlighted or not, so blocks look the same
	// whether or not the fence named a language Chroma knows.
	sb.WriteString("html pre {\n")
	sb.WriteString("    background: var(--code-bg);\n")
	sb.WriteString("    color: var(--code-fg);\n")
	sb.WriteString("    font-family: var(--code-font);\n")
	sb.WriteString("    font-size: var(--code-font-size);\n")
	sb.WriteString("    line-height: var(--code-line-height);\n")
	sb.WriteString("    font-variant-ligatures: none;\n")
	sb.WriteString("    tab-size: 4;\n")
	sb.WriteString("    margin: 4.5mm 0;\n")
	sb.WriteString("    border: none;\n")
	if highlight.PanelDisabled(cfg.Code) {
		// With no panel there is nothing for padding to sit inside, and an
		// inset would just read as an arbitrary indent.
		sb.WriteString("    padding: 0;\n")
		sb.WriteString("    border-radius: 0;\n")
	} else {
		sb.WriteString("    padding: 3.5mm 4mm;\n")
		sb.WriteString("    border-radius: 1.6mm;\n")
	}
	// Inherited text settings from body/list/figure rules would otherwise
	// justify, indent, center or hyphenate source code.
	sb.WriteString("    text-align: left;\n")
	sb.WriteString("    text-indent: 0;\n")
	sb.WriteString("    hyphens: none;\n")
	sb.WriteString("    orphans: 2;\n")
	sb.WriteString("    widows: 2;\n")
	// Keep the padding and background on both halves of a split block.
	sb.WriteString("    box-decoration-break: clone;\n")
	sb.WriteString("    -webkit-box-decoration-break: clone;\n")
	sb.WriteString("}\n\n")

	sb.WriteString("html pre code {\n")
	sb.WriteString("    display: block;\n")
	sb.WriteString("    font-family: inherit;\n")
	sb.WriteString("    font-size: inherit;\n")
	sb.WriteString("    line-height: inherit;\n")
	sb.WriteString("    color: inherit;\n")
	sb.WriteString("    background: none;\n")
	sb.WriteString("    border: none;\n")
	sb.WriteString("    border-radius: 0;\n")
	sb.WriteString("    padding: 0;\n")
	sb.WriteString("}\n\n")

	if cfg.Code.Wrap {
		// break-word only breaks a token that cannot fit on its own line,
		// unlike break-all which would split short identifiers mid-word.
		sb.WriteString("html pre, html pre code {\n")
		sb.WriteString("    white-space: pre-wrap;\n")
		sb.WriteString("    overflow-wrap: break-word;\n")
		sb.WriteString("    word-break: normal;\n")
		sb.WriteString("}\n\n")
	} else {
		sb.WriteString("html pre, html pre code { white-space: pre; }\n\n")
	}

	sb.WriteString("html pre.code-keep { break-inside: avoid; page-break-inside: avoid; }\n\n")

	// Inline code stays light on light so it reads as part of the sentence
	// instead of a one-word code block. The > guard keeps it off pre's child.
	sb.WriteString("html :not(pre) > code {\n")
	sb.WriteString("    font-family: var(--code-font);\n")
	sb.WriteString("    font-size: " + roundPx("0.9em") + ";\n")
	sb.WriteString("    background: var(--bg-soft);\n")
	sb.WriteString("    color: var(--blue);\n")
	sb.WriteString("    border-radius: 1mm;\n")
	sb.WriteString("    padding: 0.2mm 1mm;\n")
	sb.WriteString("    hyphens: none;\n")
	sb.WriteString("    overflow-wrap: break-word;\n")
	sb.WriteString("}\n\n")
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
	fmt.Fprintf(sb, "    font-size: %s;\n", roundPx(cfg.Font.Size))
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
		fmt.Fprintf(sb, "%s { font-size: %s; break-after: avoid; break-inside: avoid; page-break-after: avoid; page-break-inside: avoid;", h.tag, roundPx(h.level.Size))
		if h.level.PageBreakBefore {
			sb.WriteString(" break-before: page; page-break-before: always;")
		}
		sb.WriteString(" }\n")

		// A heading that already forces its own page break makes an "hr"
		// divider right before it redundant, and that hr is exactly the kind
		// of lone block that can get stranded on an otherwise blank page
		// (it overflows onto a fresh page by itself, then the heading's own
		// forced break bumps the heading past it to the page after). Hiding
		// it removes both the redundancy and the stranding.
		if h.level.PageBreakBefore {
			fmt.Fprintf(sb, "hr:has(+ %s) { display: none; }\n", h.tag)
		}
	}
}

// writeHeadingKeepTogetherRules keeps each heading glued to the block that
// follows it, so a heading is never stranded alone at the bottom of a page:
// break-after: avoid (above) pushes the heading forward whenever the next
// block doesn't fit at all, and break-inside: avoid here stops that next
// block from being split after just a line or two, leaving the heading with
// almost nothing under it.
func writeHeadingKeepTogetherRules(sb *strings.Builder) {
	headings := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
	followers := []string{"p", "ul", "ol", "table", "blockquote"}

	var selectors []string
	for _, h := range headings {
		for _, f := range followers {
			selectors = append(selectors, h+" + "+f)
		}
	}

	fmt.Fprintf(sb, "%s {\n", strings.Join(selectors, ",\n"))
	sb.WriteString("    break-inside: avoid;\n")
	sb.WriteString("    page-break-inside: avoid;\n")
	sb.WriteString("}\n\n")
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
	return fmt.Sprintf(`font-family: "%s", sans-serif; font-size: %s; color: var(--text-muted);`, fontFamily, roundPx("8pt"))
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
