package tmpl

import (
	"strings"
	"testing"

	"github.com/tomsiouan/templify/config"
	"github.com/tomsiouan/templify/highlight"
)

func TestBuildConfigCSS(t *testing.T) {
	t.Run("default config produces page and body rules", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		for _, want := range []string{"<style>", "</style>", ":root {", "@page {", "body {", "h1 {", "h6 {"} {
			if !strings.Contains(css, want) {
				t.Errorf("expected %q in CSS output", want)
			}
		}
	})

	t.Run("colors injected into :root", func(t *testing.T) {
		cfg := config.Default()
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, cfg.Colors.Primary) {
			t.Errorf("expected primary color %q in CSS", cfg.Colors.Primary)
		}
	})

	t.Run("font sizes are wrapped in round() to land on a whole pixel", func(t *testing.T) {
		// Chromium prints @font-face text one glyph per positioned run, and a
		// fractional pixel size makes consecutive advances inconsistent by a
		// sub-pixel amount; some PDF readers (confirmed: Apple's PDFKit) then
		// misorder text around that, most visibly around underscores in
		// monospace code. round(value, 1px) is a no-op when already whole, so
		// every font-size should carry it unconditionally.
		cfg := config.Default()
		css := string(buildConfigCSS(cfg))
		for _, want := range []string{
			"font-size: round(" + cfg.Font.Size + ", 1px);",
			"font-size: round(" + cfg.Headings.H1.Size + ", 1px);",
			"font-size: round(0.9em, 1px);", // inline code
		} {
			if !strings.Contains(css, want) {
				t.Errorf("expected %q in CSS output:\n%s", want, css)
			}
		}
	})

	t.Run("justify enabled adds text-align rule", func(t *testing.T) {
		cfg := config.Default()
		cfg.Justify = true
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, "text-align: justify") {
			t.Error("expected text-align: justify")
		}
	})

	t.Run("justify disabled omits text-align rule", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		if strings.Contains(css, "text-align: justify") {
			t.Error("unexpected text-align: justify")
		}
	})

	t.Run("paragraph indent adds text-indent rule", func(t *testing.T) {
		cfg := config.Default()
		cfg.ParagraphIndent = "1em"
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, "text-indent: 1em") {
			t.Error("expected text-indent rule")
		}
	})

	t.Run("heading indent adds padding-left rules", func(t *testing.T) {
		cfg := config.Default()
		cfg.HeadingIndent = "10px"
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, "padding-left: 10px") {
			t.Error("expected padding-left rule for h3")
		}
	})

	t.Run("heading numbers enabled adds counter rules", func(t *testing.T) {
		cfg := config.Default()
		cfg.HeadingNumbers.Enabled = true
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, "counter(h2c)") {
			t.Error("expected heading counter CSS")
		}
	})

	t.Run("headings avoid being stranded alone at the bottom of a page", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		if !strings.Contains(css, "break-after: avoid") {
			t.Error("expected break-after: avoid on headings")
		}
		if !strings.Contains(css, "h1 + p") || !strings.Contains(css, "h6 + blockquote") {
			t.Error("expected heading+follower keep-together selectors")
		}
	})

	t.Run("forced page break before a heading hides the preceding hr divider", func(t *testing.T) {
		cfg := config.Default()
		cfg.Headings.H2.PageBreakBefore = true
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, "hr:has(+ h2) { display: none; }") {
			t.Error("expected hr divider before h2 to be hidden")
		}
		if strings.Contains(css, "hr:has(+ h1)") {
			t.Error("did not expect hr rule for h1, which has no forced page break")
		}
	})

	t.Run("header background adds pagedjs margin rules", func(t *testing.T) {
		cfg := config.Default()
		cfg.Header.Enabled = true
		cfg.Header.Background = "#ffffff"
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, ".pagedjs_margin-top") {
			t.Error("expected pagedjs margin-top rule")
		}
	})

	t.Run("footer background adds pagedjs margin rules", func(t *testing.T) {
		cfg := config.Default()
		cfg.Footer.Enabled = true
		cfg.Footer.Background = "#f0f0f0"
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, ".pagedjs_margin-bottom") {
			t.Error("expected pagedjs margin-bottom rule")
		}
	})
}

func TestBuildConfigCSSCode(t *testing.T) {
	t.Run("default config styles code blocks as a filled panel", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		for _, want := range []string{
			"--code-bg: #272822;",
			"--code-fg: #f8f8f2;",
			"--code-font-size: round(8.5pt, 1px);",
			"html pre {",
			"background: var(--code-bg);",
			"html pre code {",
			"html :not(pre) > code {",
		} {
			if !strings.Contains(css, want) {
				t.Errorf("expected %q in CSS output", want)
			}
		}
	})

	t.Run("panel selectors outrank a bundle's bare pre rule", func(t *testing.T) {
		// A bundle repainting only the background would break contrast against
		// the theme's token colors, so the generated rules must win.
		css := string(buildConfigCSS(config.Default()))
		if strings.Contains(css, "\npre {") {
			t.Error("panel rule emitted without the html prefix")
		}
	})

	t.Run("inline code uses box-shadow instead of padding", func(t *testing.T) {
		// Padding on the inline <code> element made Chromium compute a
		// slightly different line-box height for the one line carrying it,
		// which some PDF readers (confirmed: Apple's PDFKit) occasionally
		// reordered relative to its neighbor on copy. box-shadow's spread
		// paints outside the box without taking part in layout, giving the
		// same visual highlight with none of that side effect.
		css := string(buildConfigCSS(config.Default()))
		start := strings.Index(css, "html :not(pre) > code {")
		if start < 0 {
			t.Fatal("inline code rule not found")
		}
		end := strings.Index(css[start:], "}")
		rule := css[start : start+end]
		if !strings.Contains(rule, "padding: 0;") {
			t.Errorf("expected zero padding on inline code:\n%s", rule)
		}
		if !strings.Contains(rule, "box-shadow:") {
			t.Errorf("expected a box-shadow standing in for padding:\n%s", rule)
		}
	})

	t.Run("code font stack keeps monospace fallbacks", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		if !strings.Contains(css, `--code-font: "JetBrains Mono", ui-monospace`) {
			t.Error("expected configured family ahead of the monospace fallbacks")
		}
	})

	t.Run("font url is imported before any rule", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		importAt := strings.Index(css, "@import url(")
		if importAt < 0 {
			t.Fatal("expected an @import for the code font")
		}
		if root := strings.Index(css, ":root {"); importAt > root {
			t.Errorf("@import at %d comes after :root at %d, so it is ignored", importAt, root)
		}
	})

	t.Run("no font url omits the import", func(t *testing.T) {
		cfg := config.Default()
		cfg.Code.FontURL = ""
		if strings.Contains(string(buildConfigCSS(cfg)), "@import") {
			t.Error("unexpected @import with no code font URL")
		}
	})

	t.Run("theme tokens are emitted when highlighting is on", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		if !strings.Contains(css, ".chroma .k") {
			t.Error("expected chroma token rules")
		}
	})

	t.Run("theme none omits token rules but keeps the panel", func(t *testing.T) {
		cfg := config.Default()
		cfg.Code.Theme = "none"
		css := string(buildConfigCSS(cfg))
		if strings.Contains(css, ".chroma .k") {
			t.Error("unexpected chroma token rules with highlighting off")
		}
		if !strings.Contains(css, "html pre {") {
			t.Error("expected the panel rule to survive with highlighting off")
		}
	})

	t.Run("background override reaches the variable", func(t *testing.T) {
		cfg := config.Default()
		cfg.Code.Background = "#3b4252"
		if !strings.Contains(string(buildConfigCSS(cfg)), "--code-bg: #3b4252;") {
			t.Error("expected the background override in --code-bg")
		}
	})

	t.Run("background none drops the panel and its padding", func(t *testing.T) {
		cfg := config.Default()
		cfg.Code.Background = "none"
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, "--code-bg: transparent;") {
			t.Error("expected a transparent --code-bg")
		}
		if !strings.Contains(css, "--code-fg: var(--text);") {
			t.Error("expected the body text color with no panel behind the code")
		}
		if strings.Contains(css, "padding: 3.5mm 4mm;") {
			t.Error("unexpected panel padding with no panel")
		}
	})

	t.Run("line numbers off omits the gutter rules", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		if strings.Contains(css, ".chroma .lnt") {
			t.Error("unexpected line number rules by default")
		}
	})

	t.Run("line numbers on emits the gutter rules", func(t *testing.T) {
		cfg := config.Default()
		cfg.Code.LineNumbers = true
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, ".chroma .ln") {
			t.Error("expected line number rules")
		}
	})

	t.Run("wrap enabled wraps long lines without splitting words", func(t *testing.T) {
		css := string(buildConfigCSS(config.Default()))
		if !strings.Contains(css, "white-space: pre-wrap;") {
			t.Error("expected pre-wrap when wrap is on")
		}
		// break-all would split short identifiers mid-word.
		if strings.Contains(css, "word-break: break-all") {
			t.Error("unexpected word-break: break-all")
		}
	})

	t.Run("wrap disabled keeps lines intact", func(t *testing.T) {
		cfg := config.Default()
		cfg.Code.Wrap = false
		css := string(buildConfigCSS(cfg))
		if !strings.Contains(css, "html pre, html pre code { white-space: pre; }") {
			t.Error("expected white-space: pre when wrap is off")
		}
	})
}

func TestCodeFontStack(t *testing.T) {
	t.Run("quotes the configured family", func(t *testing.T) {
		got := codeFontStack("JetBrains Mono")
		if !strings.HasPrefix(got, `"JetBrains Mono", `) {
			t.Errorf("got %q, want a quoted family first", got)
		}
	})

	t.Run("empty family falls back to the mono stack alone", func(t *testing.T) {
		if got := codeFontStack("  "); got != highlight.MonoFallbacks {
			t.Errorf("got %q, want %q", got, highlight.MonoFallbacks)
		}
	})
}

func TestMarginBoxStyle(t *testing.T) {
	got := marginBoxStyle("Inter")
	if !strings.Contains(got, "Inter") {
		t.Errorf("expected font family in output: %q", got)
	}
	if !strings.Contains(got, "font-family") {
		t.Errorf("expected font-family property: %q", got)
	}
	if !strings.Contains(got, "round(8pt, 1px)") {
		t.Errorf("expected the margin box size wrapped in round(): %q", got)
	}
}

func TestRoundPx(t *testing.T) {
	tests := []struct{ value, want string }{
		{"8.5pt", "round(8.5pt, 1px)"},
		{"11pt", "round(11pt, 1px)"},
		{"0.9em", "round(0.9em, 1px)"},
		{"16px", "round(16px, 1px)"},
	}
	for _, tc := range tests {
		if got := roundPx(tc.value); got != tc.want {
			t.Errorf("roundPx(%q) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestSuppressAllMarginBoxes(t *testing.T) {
	got := suppressAllMarginBoxes()
	for _, box := range []string{"@top-left", "@top-center", "@top-right", "@bottom-left", "@bottom-center", "@bottom-right"} {
		if !strings.Contains(got, box) {
			t.Errorf("expected %q in suppressed boxes", box)
		}
	}
}
