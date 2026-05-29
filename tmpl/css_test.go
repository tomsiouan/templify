package tmpl

import (
	"strings"
	"testing"

	"github.com/tomsiouan/templify/config"
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

func TestMarginBoxStyle(t *testing.T) {
	got := marginBoxStyle("Inter")
	if !strings.Contains(got, "Inter") {
		t.Errorf("expected font family in output: %q", got)
	}
	if !strings.Contains(got, "font-family") {
		t.Errorf("expected font-family property: %q", got)
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
