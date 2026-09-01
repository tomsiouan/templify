// Package highlight wires Chroma syntax highlighting into the Markdown
// pipeline and generates the matching token stylesheet, so the parser and the
// template layer agree on a single theme without either owning Chroma.
package highlight

import (
	"bytes"
	"log/slog"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"

	"github.com/tomsiouan/templify/config"
)

const (
	// ThemeNone is the config value that turns syntax highlighting off while
	// keeping the code block panel style.
	ThemeNone = "none"

	// DefaultTheme is the theme used when the configured one is unknown to
	// Chroma, and the one config.Default() ships with.
	DefaultTheme = "monokai"

	// fallbackBackground and fallbackForeground style the panel when
	// highlighting is off, or when the theme leaves those colors unset. They
	// mirror DefaultTheme so turning highlighting off only drops the colors.
	fallbackBackground = "#272822"
	fallbackForeground = "#f8f8f2"

	// MonoFallbacks is the font stack appended after the configured code font,
	// so blocks still render as monospace when that font cannot be loaded.
	MonoFallbacks = `ui-monospace, "SFMono-Regular", "SF Mono", Menlo, Consolas, "Liberation Mono", monospace`
)

// Enabled reports whether cfg asks for syntax highlighting.
func Enabled(cfg config.CodeConfig) bool {
	theme := strings.TrimSpace(cfg.Theme)
	return theme != "" && !strings.EqualFold(theme, ThemeNone)
}

// Themes returns the sorted names of every theme Chroma can render.
func Themes() []string {
	names := make([]string, 0, len(styles.Registry))
	for name := range styles.Registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Style resolves cfg.Theme to a Chroma style, warning and falling back to
// DefaultTheme when the name is unknown. Returns nil when highlighting is off.
func Style(cfg config.CodeConfig) *chroma.Style {
	if !Enabled(cfg) {
		return nil
	}
	name := strings.ToLower(strings.TrimSpace(cfg.Theme))
	if style, ok := styles.Registry[name]; ok {
		return style
	}
	slog.Warn("unknown code theme", "theme", cfg.Theme, "fallback", DefaultTheme)
	return styles.Registry[DefaultTheme]
}

// Extension returns the goldmark extension that highlights fenced code blocks,
// or nil when highlighting is off. Token colors are emitted as classes rather
// than inline styles so CSS from the config keeps the last word.
func Extension(cfg config.CodeConfig) goldmark.Extender {
	style := Style(cfg)
	if style == nil {
		return nil
	}
	return highlighting.NewHighlighting(
		highlighting.WithCustomStyle(style),
		highlighting.WithFormatOptions(
			chromahtml.WithClasses(true),
			chromahtml.WithLineNumbers(cfg.LineNumbers),
		),
	)
}

// PanelDisabled reports whether cfg turns the filled panel off, leaving code on
// the page background.
func PanelDisabled(cfg config.CodeConfig) bool {
	switch strings.ToLower(strings.TrimSpace(cfg.Background)) {
	case "none", "transparent":
		return true
	}
	return false
}

// Palette returns the code block background and foreground colors: explicit
// config values win, then the theme's own colors, then the fallbacks.
//
// With the panel off, the theme's own text color is dropped unless it was set
// explicitly: it was picked to sit on the theme's background, not on the page.
func Palette(cfg config.CodeConfig) (background, foreground string) {
	if PanelDisabled(cfg) {
		foreground = "var(--text)"
		if c := strings.TrimSpace(cfg.Foreground); c != "" {
			foreground = c
		}
		warnIfPanelNeeded(cfg)
		return "transparent", foreground
	}

	background, foreground = fallbackBackground, fallbackForeground
	if style := Style(cfg); style != nil {
		entry := style.Get(chroma.Background)
		if entry.Background.IsSet() {
			background = entry.Background.String()
		}
		if entry.Colour.IsSet() {
			foreground = entry.Colour.String()
		}
	}
	if c := strings.TrimSpace(cfg.Background); c != "" {
		background = c
	}
	if c := strings.TrimSpace(cfg.Foreground); c != "" {
		foreground = c
	}
	return background, foreground
}

// warnIfPanelNeeded flags a dark theme left on a page with no panel behind it.
// The token colors were chosen against the theme's dark background, so on white
// paper they wash out; only the panel was making them readable.
func warnIfPanelNeeded(cfg config.CodeConfig) {
	style := Style(cfg)
	if style == nil {
		return
	}
	bg := style.Get(chroma.Background).Background
	if bg.IsSet() && bg.Brightness() < 0.5 {
		slog.Warn("dark code theme with no panel background may be unreadable on white",
			"theme", cfg.Theme, "hint", "use a light theme or theme: none")
	}
}

// panelRules are the Chroma rules that paint the block itself. They are dropped
// so the generated document stylesheet is the only place the panel is styled,
// which keeps the Background/Foreground config overrides authoritative.
var panelRules = map[string]bool{
	"Background": true,
	"PreWrapper": true,
}

// gutterRules are the Chroma rules that only matter once line numbers are on.
// "Line" is the important one: it turns every line into a flex container, which
// interferes with wrapping and page breaking when there is no gutter to align.
var gutterRules = map[string]bool{
	"Line":             true,
	"LineHighlight":    true,
	"LineLink":         true,
	"LineNumbers":      true,
	"LineNumbersTable": true,
	"LineTable":        true,
	"LineTableTD":      true,
}

// CSS returns the Chroma token stylesheet for cfg's theme, one rule per line,
// or "" when highlighting is off.
func CSS(cfg config.CodeConfig) string {
	style := Style(cfg)
	if style == nil {
		return ""
	}
	var raw bytes.Buffer
	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.WithLineNumbers(cfg.LineNumbers),
	)
	if err := formatter.WriteCSS(&raw, style); err != nil {
		slog.Warn("generate code theme CSS failed", "theme", cfg.Theme, "err", err)
		return ""
	}

	var sb strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(raw.String()), "\n") {
		name := ruleName(line)
		if panelRules[name] || (!cfg.LineNumbers && gutterRules[name]) {
			continue
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

// ruleName extracts the token name from the `/* Name */ .chroma .x { … }`
// comment Chroma prefixes each generated rule with, or "" if there is none.
func ruleName(line string) string {
	rest, ok := strings.CutPrefix(strings.TrimSpace(line), "/*")
	if !ok {
		return ""
	}
	end := strings.Index(rest, "*/")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:end])
}
