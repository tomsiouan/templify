package fontface

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tomsiouan/templify/config"
)

// buildWoff2 constructs a minimal, structurally valid (but not renderable)
// WOFF2 byte buffer with a single table directory entry naming tag via the
// "arbitrary tag" encoding (table index 0x3f), which is how isVariableFont
// reads an actual "fvar" entry. It exists to test the table-directory walk in
// isolation, without needing a real font file on disk.
func buildWoff2(t *testing.T, tag string) []byte {
	t.Helper()
	if len(tag) != 4 {
		t.Fatalf("tag must be 4 bytes, got %q", tag)
	}
	header := make([]byte, 48)
	copy(header[0:4], "wOF2")
	binary.BigEndian.PutUint16(header[12:14], 1) // numTables = 1

	entry := []byte{0x3f} // flags: arbitrary tag, no transform
	entry = append(entry, []byte(tag)...)
	entry = append(entry, 0x0A) // origLength, UIntBase128-encoded 10

	return append(header, entry...)
}

// buildWoff2KnownTag is buildWoff2's counterpart for a table referenced by its
// index into woff2KnownTags rather than spelled out inline.
func buildWoff2KnownTag(t *testing.T, idx byte) []byte {
	t.Helper()
	header := make([]byte, 48)
	copy(header[0:4], "wOF2")
	binary.BigEndian.PutUint16(header[12:14], 1)

	entry := []byte{idx & 0x3f, 0x0A}
	return append(header, entry...)
}

// writeFakeWoff2 writes a minimal WOFF2 file at path with a single, ordinary
// "cmap" table. It exists purely to exercise the file-format plumbing in
// BuildCSS (extension check, data URI encoding) where variable-ness is
// irrelevant; isVariableFont itself is tested directly in fetch_test.go.
func writeFakeWoff2(t *testing.T, path string) {
	t.Helper()
	data := buildWoff2(t, "cmap")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func loadConfigAt(t *testing.T, dir string) *config.Config {
	t.Helper()
	path := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(path, []byte("justify: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestBuildCSSNoFaces(t *testing.T) {
	cfg := config.Default()
	css, err := BuildCSS(cfg)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if css != "" {
		t.Errorf("got %q, want empty with no faces configured", css)
	}
}

func TestBuildCSSFontFace(t *testing.T) {
	dir := t.TempDir()
	cfg := loadConfigAt(t, dir)
	cfg.Font.Family = "Inter"

	fontPath := filepath.Join(dir, "inter-400.woff2")
	writeFakeWoff2(t, fontPath)
	cfg.Font.Faces = []config.FontFace{
		{File: "./inter-400.woff2", Weight: 400, UnicodeRange: "U+0000-00FF"},
	}

	css, err := BuildCSS(cfg)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	for _, want := range []string{
		"@font-face {",
		`font-family: "Inter"`,
		"font-weight: 400",
		"font-style: normal", // defaulted
		"src: url(data:font/woff2;base64,",
		"unicode-range: U+0000-00FF",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("expected %q in:\n%s", want, css)
		}
	}
}

func TestBuildCSSDefaultsWeightAndStyle(t *testing.T) {
	dir := t.TempDir()
	cfg := loadConfigAt(t, dir)
	cfg.Font.Family = "Inter"
	writeFakeWoff2(t, filepath.Join(dir, "inter.woff2"))
	cfg.Font.Faces = []config.FontFace{{File: "./inter.woff2"}} // no weight, no style

	css, err := BuildCSS(cfg)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(css, "font-weight: 400") {
		t.Errorf("expected default weight 400 in:\n%s", css)
	}
	if !strings.Contains(css, "font-style: normal") {
		t.Errorf("expected default style normal in:\n%s", css)
	}
	if strings.Contains(css, "unicode-range") {
		t.Errorf("unexpected unicode-range with none configured:\n%s", css)
	}
}

func TestBuildCSSBothFontAndCode(t *testing.T) {
	dir := t.TempDir()
	cfg := loadConfigAt(t, dir)
	cfg.Font.Family = "Inter"
	cfg.Code.FontFamily = "JetBrains Mono"
	writeFakeWoff2(t, filepath.Join(dir, "inter.woff2"))
	writeFakeWoff2(t, filepath.Join(dir, "mono.woff2"))
	cfg.Font.Faces = []config.FontFace{{File: "./inter.woff2", Weight: 400}}
	cfg.Code.Faces = []config.FontFace{{File: "./mono.woff2", Weight: 400}}

	css, err := BuildCSS(cfg)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if strings.Count(css, "@font-face") != 2 {
		t.Errorf("expected 2 @font-face rules, got:\n%s", css)
	}
	if !strings.Contains(css, `font-family: "Inter"`) || !strings.Contains(css, `font-family: "JetBrains Mono"`) {
		t.Errorf("expected both families in:\n%s", css)
	}
}

func TestBuildCSSErrors(t *testing.T) {
	t.Run("faces with empty family", func(t *testing.T) {
		dir := t.TempDir()
		cfg := loadConfigAt(t, dir)
		writeFakeWoff2(t, filepath.Join(dir, "f.woff2"))
		cfg.Font.Family = "  "
		cfg.Font.Faces = []config.FontFace{{File: "./f.woff2"}}

		if _, err := BuildCSS(cfg); err == nil {
			t.Error("expected an error for an empty family with faces configured")
		}
	})

	t.Run("face with empty file", func(t *testing.T) {
		dir := t.TempDir()
		cfg := loadConfigAt(t, dir)
		cfg.Font.Family = "Inter"
		cfg.Font.Faces = []config.FontFace{{File: "  "}}

		if _, err := BuildCSS(cfg); err == nil {
			t.Error("expected an error for a face with no file")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		dir := t.TempDir()
		cfg := loadConfigAt(t, dir)
		cfg.Font.Family = "Inter"
		cfg.Font.Faces = []config.FontFace{{File: "./does-not-exist.woff2"}}

		if _, err := BuildCSS(cfg); err == nil {
			t.Error("expected an error for a missing font file")
		}
	})

	t.Run("unsupported extension", func(t *testing.T) {
		dir := t.TempDir()
		cfg := loadConfigAt(t, dir)
		cfg.Font.Family = "Inter"
		path := filepath.Join(dir, "inter.eot")
		if err := os.WriteFile(path, []byte("not a font"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg.Font.Faces = []config.FontFace{{File: "./inter.eot"}}

		if _, err := BuildCSS(cfg); err == nil {
			t.Error("expected an error for an unsupported font format")
		}
	})
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		family string
		want   string
	}{
		{"Inter", "inter"},
		{"JetBrains Mono", "jetbrains-mono"},
		{"  Spaced  ", "spaced"},
		{"", "font"},
		{"   ", "font"},
		{"Font/Weird*Name!", "font-weird-name"},
	}
	for _, tc := range tests {
		t.Run(tc.family, func(t *testing.T) {
			if got := Prefix(tc.family); got != tc.want {
				t.Errorf("Prefix(%q) = %q, want %q", tc.family, got, tc.want)
			}
		})
	}
}

func TestStyleSuffix(t *testing.T) {
	tests := []struct{ style, want string }{
		{"", ""},
		{"normal", ""},
		{"italic", "-italic"},
		{"oblique", "-oblique"},
	}
	for _, tc := range tests {
		if got := styleSuffix(tc.style); got != tc.want {
			t.Errorf("styleSuffix(%q) = %q, want %q", tc.style, got, tc.want)
		}
	}
}

func TestIsGoogleFontsHost(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://fonts.googleapis.com/css2?family=Inter", true},
		{"https://fonts.googleapis.com/css?family=Inter", true},
		{"http://fonts.googleapis.com/css2?family=Inter", true},
		{"https://fonts.gstatic.com/s/inter/v20/foo.woff2", false},
		{"https://example.com/fonts.css", false},
		{"not a url at all \x7f", false},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			if got := isGoogleFontsHost(tc.url); got != tc.want {
				t.Errorf("isGoogleFontsHost(%q) = %v, want %v", tc.url, got, tc.want)
			}
		})
	}
}

func TestGoogleFontsSingleWeightURL(t *testing.T) {
	tests := []struct {
		family, style string
		weight        int
		want          string
	}{
		{"Inter", "normal", 600, "https://fonts.googleapis.com/css2?family=Inter:wght@600&display=swap"},
		{"JetBrains Mono", "normal", 400, "https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400&display=swap"},
		{"Inter", "italic", 400, "https://fonts.googleapis.com/css2?family=Inter:ital,wght@1,400&display=swap"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := googleFontsSingleWeightURL(tc.family, tc.weight, tc.style); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseFontFaceBlocks(t *testing.T) {
	css := `
/* latin */
@font-face {
  font-family: 'Inter';
  font-style: normal;
  font-weight: 400;
  font-display: swap;
  src: url(https://fonts.gstatic.com/s/inter/v20/latin-400.woff2) format('woff2');
  unicode-range: U+0000-00FF, U+0131;
}
/* latin-ext */
@font-face {
  font-family: 'Inter';
  font-style: italic;
  font-weight: 600;
  src: url(https://fonts.gstatic.com/s/inter/v20/ext-600i.woff2) format('woff2');
}
@font-face {
  font-family: 'Already';
  src: url(data:font/woff2;base64,AAAA) format('woff2');
}
`
	blocks := parseFontFaceBlocks(css)
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2 (data: URI block should be skipped)", len(blocks))
	}

	b0 := blocks[0]
	if b0.family != "Inter" || b0.weight != 400 || b0.style != "normal" {
		t.Errorf("block 0 = %+v", b0)
	}
	if b0.unicodeRange != "U+0000-00FF, U+0131" {
		t.Errorf("block 0 unicodeRange = %q", b0.unicodeRange)
	}
	if !strings.HasSuffix(b0.srcURL, "latin-400.woff2") {
		t.Errorf("block 0 srcURL = %q", b0.srcURL)
	}

	b1 := blocks[1]
	if b1.weight != 600 || b1.style != "italic" {
		t.Errorf("block 1 = %+v", b1)
	}
	if b1.unicodeRange != "" {
		t.Errorf("block 1 unicodeRange = %q, want empty", b1.unicodeRange)
	}
}

func TestParseFontFaceBlocksNoUsableBlocks(t *testing.T) {
	css := `@font-face { font-family: 'X'; src: url(data:font/woff2;base64,AAAA) format('woff2'); }`
	blocks := parseFontFaceBlocks(css)
	if len(blocks) != 0 {
		t.Errorf("got %d blocks, want 0", len(blocks))
	}
}

func TestFacesYAML(t *testing.T) {
	fetched := []Fetched{
		{Path: "/abs/project/fonts/inter-400.woff2", Weight: 400},
		{Path: "/abs/project/fonts/inter-600i.woff2", Weight: 600, Style: "italic", UnicodeRange: "U+0000-00FF"},
	}
	got := FacesYAML("font", fetched, "/abs/project")

	for _, want := range []string{
		"font:\n  faces:\n",
		"- file: ./fonts/inter-400.woff2",
		"weight: 400",
		"- file: ./fonts/inter-600i.woff2",
		"style: italic",
		`unicode_range: "U+0000-00FF"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in:\n%s", want, got)
		}
	}
	// The normal-style, no-range face should not grow extra fields.
	if strings.Contains(got, "style: normal") {
		t.Errorf("unexpected explicit normal style in:\n%s", got)
	}
}

func TestFacesYAMLPathNotRelative(t *testing.T) {
	// filepath.Rel requires both paths to be absolute or both relative; mixing
	// them, as here, is the simplest portable way to make it fail.
	fetched := []Fetched{{Path: "relative/inter.woff2", Weight: 400}}
	got := FacesYAML("font", fetched, "/abs/project")
	if !strings.Contains(got, "file: relative/inter.woff2") {
		t.Errorf("expected the raw path kept as-is when it cannot be made relative:\n%s", got)
	}
}
