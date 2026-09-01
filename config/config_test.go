package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tomsiouan/templify/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.Page.Size != "A4" {
		t.Errorf("Page.Size = %q, want A4", cfg.Page.Size)
	}
	if cfg.Font.Family != "Inter" {
		t.Errorf("Font.Family = %q, want Inter", cfg.Font.Family)
	}
	if !cfg.TOC.Enabled {
		t.Error("TOC should be enabled by default")
	}
	if cfg.Header.Enabled {
		t.Error("header should be disabled by default")
	}
	if !cfg.Footer.Enabled {
		t.Error("footer should be enabled by default")
	}
	if cfg.Code.Theme != "monokai" {
		t.Errorf("Code.Theme = %q, want monokai", cfg.Code.Theme)
	}
	if !cfg.Code.Wrap {
		t.Error("code wrapping should be enabled by default")
	}
	if cfg.Code.LineNumbers {
		t.Error("code line numbers should be disabled by default")
	}
	if cfg.Code.KeepTogetherLines <= 0 {
		t.Errorf("Code.KeepTogetherLines = %d, want a positive default", cfg.Code.KeepTogetherLines)
	}
}

func TestCodeConfigMerge(t *testing.T) {
	t.Run("overrides the named keys and keeps the rest", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "cfg.yml")
		body := "code:\n  theme: monokai\n  line_numbers: true\n  keep_together_lines: 40\n"
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if cfg.Code.Theme != "monokai" {
			t.Errorf("Code.Theme = %q, want monokai", cfg.Code.Theme)
		}
		if !cfg.Code.LineNumbers {
			t.Error("Code.LineNumbers should be true")
		}
		if cfg.Code.KeepTogetherLines != 40 {
			t.Errorf("Code.KeepTogetherLines = %d, want 40", cfg.Code.KeepTogetherLines)
		}
		if cfg.Code.FontSize != "8.5pt" {
			t.Errorf("Code.FontSize = %q, want the default 8.5pt", cfg.Code.FontSize)
		}
	})

	t.Run("keep_together_lines zero disables keeping blocks whole", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "cfg.yml")
		if err := os.WriteFile(path, []byte("code:\n  keep_together_lines: 0\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if cfg.Code.KeepTogetherLines != 0 {
			t.Errorf("Code.KeepTogetherLines = %d, want 0", cfg.Code.KeepTogetherLines)
		}
	})
}

func TestFontFacesMerge(t *testing.T) {
	t.Run("parses a faces list for the body and code fonts", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "cfg.yml")
		body := `
font:
  family: Inter
  faces:
    - file: ./fonts/inter-400.woff2
      weight: 400
    - file: ./fonts/inter-600.woff2
      weight: 600
      unicode_range: "U+0000-00FF"
code:
  faces:
    - file: ./fonts/mono-400.woff2
      style: italic
`
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("error: %v", err)
		}

		if len(cfg.Font.Faces) != 2 {
			t.Fatalf("Font.Faces len = %d, want 2", len(cfg.Font.Faces))
		}
		if cfg.Font.Faces[0].File != "./fonts/inter-400.woff2" || cfg.Font.Faces[0].Weight != 400 {
			t.Errorf("Font.Faces[0] = %+v", cfg.Font.Faces[0])
		}
		if cfg.Font.Faces[1].UnicodeRange != "U+0000-00FF" {
			t.Errorf("Font.Faces[1].UnicodeRange = %q, want U+0000-00FF", cfg.Font.Faces[1].UnicodeRange)
		}

		if len(cfg.Code.Faces) != 1 {
			t.Fatalf("Code.Faces len = %d, want 1", len(cfg.Code.Faces))
		}
		if cfg.Code.Faces[0].Style != "italic" {
			t.Errorf("Code.Faces[0].Style = %q, want italic", cfg.Code.Faces[0].Style)
		}
	})

	t.Run("faces default to empty, url is untouched", func(t *testing.T) {
		cfg := config.Default()
		if len(cfg.Font.Faces) != 0 {
			t.Errorf("Font.Faces = %+v, want empty by default", cfg.Font.Faces)
		}
		if cfg.Font.URL == "" {
			t.Error("Font.URL should keep its default when no faces are configured")
		}
	})
}

func TestFontFacesCSS(t *testing.T) {
	cfg := config.Default()
	if got := cfg.FontFacesCSS(); got != "" {
		t.Errorf("FontFacesCSS() = %q, want empty before SetFontFacesCSS", got)
	}
	cfg.SetFontFacesCSS("@font-face { font-family: 'X'; }")
	if got := cfg.FontFacesCSS(); got != "@font-face { font-family: 'X'; }" {
		t.Errorf("FontFacesCSS() = %q, want the stored value", got)
	}
}

func TestConfigDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(path, []byte("justify: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if cfg.Dir() != dir {
		t.Errorf("Dir() = %q, want %q", cfg.Dir(), dir)
	}
}

func TestContentCSS(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", `""`},
		{"hello", `"hello"`},
		{"{page}", `counter(page)`},
		{"{pages}", `counter(pages)`},
		{"Page {page}", `"Page " counter(page)`},
		{"{page} / {pages}", `counter(page) " / " counter(pages)`},
		{"Page {page} of {pages}", `"Page " counter(page) " of " counter(pages)`},
	}
	for _, tc := range tests {
		if got := config.ContentCSS(tc.input); got != tc.want {
			t.Errorf("ContentCSS(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestCustomString(t *testing.T) {
	tests := []struct {
		name   string
		custom map[string]any
		path   string
		def    string
		want   string
	}{
		{"nil custom", nil, "key", "default", "default"},
		{"missing key", map[string]any{"other": "x"}, "key", "default", "default"},
		{"flat key found", map[string]any{"key": "value"}, "key", "default", "value"},
		{"nested key found", map[string]any{"a": map[string]any{"b": "deep"}}, "a.b", "default", "deep"},
		{"type mismatch int", map[string]any{"key": 42}, "key", "default", "default"},
		{"missing nested key", map[string]any{"a": map[string]any{"c": "x"}}, "a.b", "default", "default"},
		{"non-map nested", map[string]any{"a": "notamap"}, "a.b", "default", "default"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{Custom: tc.custom}
			if got := cfg.CustomString(tc.path, tc.def); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCustomBool(t *testing.T) {
	tests := []struct {
		name   string
		custom map[string]any
		path   string
		def    bool
		want   bool
	}{
		{"nil custom", nil, "key", false, false},
		{"missing key", map[string]any{"other": true}, "key", false, false},
		{"flat true", map[string]any{"key": true}, "key", false, true},
		{"flat false explicit", map[string]any{"key": false}, "key", true, false},
		{"nested true", map[string]any{"a": map[string]any{"b": true}}, "a.b", false, true},
		{"type mismatch string", map[string]any{"key": "true"}, "key", false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{Custom: tc.custom}
			if got := cfg.CustomBool(tc.path, tc.def); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestResolvePath(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(f, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("absolute path unchanged", func(t *testing.T) {
		got := cfg.ResolvePath("/abs/file.txt")
		if got != "/abs/file.txt" {
			t.Errorf("got %q, want /abs/file.txt", got)
		}
	})

	t.Run("relative path resolved against config dir", func(t *testing.T) {
		got := cfg.ResolvePath("file.txt")
		want := filepath.Join(dir, "file.txt")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("relative nested path resolved", func(t *testing.T) {
		got := cfg.ResolvePath("sub/file.txt")
		want := filepath.Join(dir, "sub/file.txt")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestApplyDocMeta(t *testing.T) {
	t.Run("nil meta no change", func(t *testing.T) {
		cfg := config.Default()
		cfg.ApplyDocMeta(nil)
		if cfg.Header.Left != "" || cfg.Footer.Right != "{page} / {pages}" {
			t.Error("nil meta should not change config")
		}
	})

	t.Run("override header slots", func(t *testing.T) {
		cfg := config.Default()
		cfg.ApplyDocMeta(map[string]any{
			"header": map[string]any{
				"left":   "Left Header",
				"center": "Center Header",
				"right":  "Right Header",
			},
		})
		if cfg.Header.Left != "Left Header" {
			t.Errorf("Header.Left = %q", cfg.Header.Left)
		}
		if cfg.Header.Center != "Center Header" {
			t.Errorf("Header.Center = %q", cfg.Header.Center)
		}
		if cfg.Header.Right != "Right Header" {
			t.Errorf("Header.Right = %q", cfg.Header.Right)
		}
	})

	t.Run("partial footer override preserves defaults", func(t *testing.T) {
		cfg := config.Default()
		cfg.ApplyDocMeta(map[string]any{
			"footer": map[string]any{
				"left": "Footer Left",
			},
		})
		if cfg.Footer.Left != "Footer Left" {
			t.Errorf("Footer.Left = %q", cfg.Footer.Left)
		}
		if cfg.Footer.Right != "{page} / {pages}" {
			t.Errorf("Footer.Right should be default, got %q", cfg.Footer.Right)
		}
	})

	t.Run("wrong type ignored", func(t *testing.T) {
		cfg := config.Default()
		cfg.ApplyDocMeta(map[string]any{
			"header": "not a map",
		})
		if cfg.Header.Left != "" {
			t.Errorf("expected no change, got Header.Left = %q", cfg.Header.Left)
		}
	})
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(f, []byte("page:\n  size: Letter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(f)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Page.Size != "Letter" {
		t.Errorf("Page.Size = %q, want Letter", cfg.Page.Size)
	}
	if got := cfg.ResolvePath("file.txt"); got != filepath.Join(dir, "file.txt") {
		t.Errorf("ResolvePath = %q, want %q", got, filepath.Join(dir, "file.txt"))
	}
}

func TestFromBundle(t *testing.T) {
	t.Run("nil data returns default", func(t *testing.T) {
		cfg, err := config.FromBundle(nil)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Page.Size != "A4" {
			t.Errorf("expected default A4, got %q", cfg.Page.Size)
		}
	})

	t.Run("valid yaml overrides default", func(t *testing.T) {
		cfg, err := config.FromBundle([]byte("page:\n  size: Letter\n"))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Page.Size != "Letter" {
			t.Errorf("got %q, want Letter", cfg.Page.Size)
		}
	})
}

func TestMergeFile(t *testing.T) {
	t.Run("overlays yaml onto existing config", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "overlay.yml")
		if err := os.WriteFile(f, []byte("page:\n  size: Letter\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg := config.Default()
		if err := config.MergeFile(cfg, f); err != nil {
			t.Fatal(err)
		}
		if cfg.Page.Size != "Letter" {
			t.Errorf("Page.Size = %q, want Letter", cfg.Page.Size)
		}
		if cfg.Font.Family != "Inter" {
			t.Errorf("Font.Family changed unexpectedly: %q", cfg.Font.Family)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		cfg := config.Default()
		if err := config.MergeFile(cfg, "/nonexistent.yml"); err == nil {
			t.Error("expected error for missing file")
		}
	})
}
