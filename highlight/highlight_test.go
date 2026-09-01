package highlight

import (
	"strings"
	"testing"

	"github.com/tomsiouan/templify/config"
)

func TestEnabled(t *testing.T) {
	tests := []struct {
		name  string
		theme string
		want  bool
	}{
		{"empty theme disabled", "", false},
		{"none disabled", "none", false},
		{"none case insensitive", "None", false},
		{"whitespace only disabled", "   ", false},
		{"named theme enabled", "nord", true},
		{"unknown name still enabled", "not-a-theme", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Enabled(config.CodeConfig{Theme: tc.theme}); got != tc.want {
				t.Errorf("Enabled(%q) = %v, want %v", tc.theme, got, tc.want)
			}
		})
	}
}

func TestStyle(t *testing.T) {
	t.Run("nil when disabled", func(t *testing.T) {
		if got := Style(config.CodeConfig{Theme: ThemeNone}); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("resolves known theme", func(t *testing.T) {
		style := Style(config.CodeConfig{Theme: "monokai"})
		if style == nil {
			t.Fatal("got nil style")
		}
		if !strings.EqualFold(style.Name, "monokai") {
			t.Errorf("Name = %q, want monokai", style.Name)
		}
	})

	t.Run("is case insensitive", func(t *testing.T) {
		if style := Style(config.CodeConfig{Theme: "MonoKai"}); style == nil || !strings.EqualFold(style.Name, "monokai") {
			t.Errorf("got %v, want monokai", style)
		}
	})

	t.Run("falls back for unknown theme", func(t *testing.T) {
		style := Style(config.CodeConfig{Theme: "not-a-real-theme"})
		if style == nil {
			t.Fatal("got nil style")
		}
		if !strings.EqualFold(style.Name, DefaultTheme) {
			t.Errorf("Name = %q, want %s", style.Name, DefaultTheme)
		}
	})
}

func TestThemes(t *testing.T) {
	themes := Themes()
	if len(themes) == 0 {
		t.Fatal("no themes returned")
	}
	if !sortedContains(themes, DefaultTheme) {
		t.Errorf("%s missing from %d themes", DefaultTheme, len(themes))
	}
	for i := 1; i < len(themes); i++ {
		if themes[i-1] > themes[i] {
			t.Fatalf("not sorted at %d: %q > %q", i, themes[i-1], themes[i])
		}
	}
}

func sortedContains(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

func TestPalette(t *testing.T) {
	t.Run("uses theme colors", func(t *testing.T) {
		bg, fg := Palette(config.CodeConfig{Theme: "nord"})
		if bg != "#2e3440" {
			t.Errorf("background = %q, want #2e3440", bg)
		}
		if fg != "#d8dee9" {
			t.Errorf("foreground = %q, want #d8dee9", fg)
		}
	})

	t.Run("config overrides win", func(t *testing.T) {
		bg, fg := Palette(config.CodeConfig{Theme: "nord", Background: "#101010", Foreground: "#eeeeee"})
		if bg != "#101010" || fg != "#eeeeee" {
			t.Errorf("got %q/%q, want #101010/#eeeeee", bg, fg)
		}
	})

	t.Run("falls back when highlighting is off", func(t *testing.T) {
		bg, fg := Palette(config.CodeConfig{Theme: ThemeNone})
		if bg != fallbackBackground || fg != fallbackForeground {
			t.Errorf("got %q/%q, want %q/%q", bg, fg, fallbackBackground, fallbackForeground)
		}
	})

	t.Run("overrides apply with highlighting off", func(t *testing.T) {
		bg, _ := Palette(config.CodeConfig{Theme: ThemeNone, Background: "#3b4252"})
		if bg != "#3b4252" {
			t.Errorf("background = %q, want #3b4252", bg)
		}
	})

	t.Run("panel off drops the theme's text color for the body color", func(t *testing.T) {
		// monokai's #f8f8f2 was picked to sit on its own dark background, so
		// keeping it would leave near-white code on white paper.
		bg, fg := Palette(config.CodeConfig{Theme: "monokai", Background: "none"})
		if bg != "transparent" {
			t.Errorf("background = %q, want transparent", bg)
		}
		if fg != "var(--text)" {
			t.Errorf("foreground = %q, want var(--text)", fg)
		}
	})

	t.Run("panel off still honours an explicit foreground", func(t *testing.T) {
		_, fg := Palette(config.CodeConfig{Theme: "github", Background: "none", Foreground: "#222222"})
		if fg != "#222222" {
			t.Errorf("foreground = %q, want #222222", fg)
		}
	})
}

func TestPanelDisabled(t *testing.T) {
	tests := []struct {
		background string
		want       bool
	}{
		{"", false},
		{"#272822", false},
		{"none", true},
		{"transparent", true},
		{"None", true},
		{"  TRANSPARENT  ", true},
	}
	for _, tc := range tests {
		t.Run(tc.background, func(t *testing.T) {
			if got := PanelDisabled(config.CodeConfig{Background: tc.background}); got != tc.want {
				t.Errorf("PanelDisabled(%q) = %v, want %v", tc.background, got, tc.want)
			}
		})
	}
}

func TestExtension(t *testing.T) {
	if got := Extension(config.CodeConfig{Theme: ThemeNone}); got != nil {
		t.Errorf("got %v, want nil when disabled", got)
	}
	if got := Extension(config.CodeConfig{Theme: "nord"}); got == nil {
		t.Error("got nil extension for an enabled theme")
	}
}

func TestCSS(t *testing.T) {
	t.Run("empty when disabled", func(t *testing.T) {
		if got := CSS(config.CodeConfig{Theme: ThemeNone}); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("emits token rules", func(t *testing.T) {
		css := CSS(config.CodeConfig{Theme: "nord"})
		if !strings.Contains(css, ".chroma .k ") && !strings.Contains(css, ".chroma .k{") {
			t.Errorf("keyword rule missing from:\n%s", css)
		}
	})

	t.Run("drops the panel rules", func(t *testing.T) {
		css := CSS(config.CodeConfig{Theme: "nord"})
		// Both would repaint the block and defeat the Background override.
		if strings.Contains(css, "/* Background */") || strings.Contains(css, "/* PreWrapper */") {
			t.Errorf("panel rules leaked into:\n%s", css)
		}
	})

	t.Run("drops gutter rules without line numbers", func(t *testing.T) {
		css := CSS(config.CodeConfig{Theme: "nord"})
		// The Line rule turns each line into a flex container, which is only
		// wanted when there are numbers to align against.
		if strings.Contains(css, "/* Line */") {
			t.Errorf("Line rule kept without line numbers:\n%s", css)
		}
	})

	t.Run("keeps gutter rules with line numbers", func(t *testing.T) {
		css := CSS(config.CodeConfig{Theme: "nord", LineNumbers: true})
		for _, want := range []string{"/* Line */", "/* LineNumbers */"} {
			if !strings.Contains(css, want) {
				t.Errorf("%s missing from:\n%s", want, css)
			}
		}
	})
}

func TestRuleName(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{"named rule", "/* Keyword */ .chroma .k { color: #81a1c1 }", "Keyword"},
		{"leading whitespace", "  /* Background */ .bg { color: red }", "Background"},
		{"no comment", ".chroma .k { color: red }", ""},
		{"unterminated comment", "/* Keyword .chroma .k {}", ""},
		{"empty line", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ruleName(tc.line); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
