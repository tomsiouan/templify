package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tomsiouan/templify/config"
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns
// everything written to it. fetchFonts prints its `faces:` blocks with
// fmt.Println/fmt.Print, which always target os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestInlineFontsNoop(t *testing.T) {
	t.Run("no faces, no url", func(t *testing.T) {
		cfg := config.Default()
		cfg.Font.URL = ""
		cfg.Code.FontURL = ""

		if err := inlineFonts(cfg); err != nil {
			t.Fatalf("error: %v", err)
		}
		if cfg.FontFacesCSS() != "" {
			t.Errorf("FontFacesCSS() = %q, want empty", cfg.FontFacesCSS())
		}
	})

	t.Run("url present but no faces leaves url untouched", func(t *testing.T) {
		cfg := config.Default() // ships a default Font.URL and Code.FontURL
		wantFontURL := cfg.Font.URL
		wantCodeURL := cfg.Code.FontURL

		if err := inlineFonts(cfg); err != nil {
			t.Fatalf("error: %v", err)
		}
		if cfg.Font.URL != wantFontURL {
			t.Errorf("Font.URL = %q, want unchanged %q", cfg.Font.URL, wantFontURL)
		}
		if cfg.Code.FontURL != wantCodeURL {
			t.Errorf("Code.FontURL = %q, want unchanged %q", cfg.Code.FontURL, wantCodeURL)
		}
		if cfg.FontFacesCSS() != "" {
			t.Errorf("FontFacesCSS() = %q, want empty with no faces configured", cfg.FontFacesCSS())
		}
	})
}

func TestInlineFontsWithFaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(path, []byte("justify: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Font.Family = "Inter"
	fontPath := filepath.Join(dir, "inter.woff2")
	if err := os.WriteFile(fontPath, buildTestWoff2(t), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg.Font.Faces = []config.FontFace{{File: "./inter.woff2", Weight: 400}}

	if err := inlineFonts(cfg); err != nil {
		t.Fatalf("error: %v", err)
	}

	if cfg.FontFacesCSS() == "" {
		t.Error("expected FontFacesCSS() to be populated")
	}
	if !strings.Contains(cfg.FontFacesCSS(), `font-family: "Inter"`) {
		t.Errorf("expected Inter in FontFacesCSS():\n%s", cfg.FontFacesCSS())
	}
	if cfg.Font.URL != "" {
		t.Errorf("Font.URL = %q, want cleared once faces are inlined", cfg.Font.URL)
	}
}

func TestInlineFontsCodeFacesDoNotClearFontURL(t *testing.T) {
	// The code font's <link> suppression happens in the CSS generator
	// (writeCodeFontImport checks len(Code.Faces) directly); inlineFonts only
	// needs to blank the body font's URL, since that one drives a bundle's own
	// <link> tag rather than something the CSS generator controls.
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(path, []byte("justify: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Code.FontFamily = "Mono"
	monoPath := filepath.Join(dir, "mono.woff2")
	if err := os.WriteFile(monoPath, buildTestWoff2(t), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg.Code.Faces = []config.FontFace{{File: "./mono.woff2", Weight: 400}}
	wantCodeURL := cfg.Code.FontURL

	if err := inlineFonts(cfg); err != nil {
		t.Fatalf("error: %v", err)
	}
	if cfg.Code.FontURL != wantCodeURL {
		t.Errorf("Code.FontURL = %q, want unchanged %q", cfg.Code.FontURL, wantCodeURL)
	}
}

func TestInlineFontsPropagatesBuildError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(path, []byte("justify: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Font.Family = "Inter"
	cfg.Font.Faces = []config.FontFace{{File: "./does-not-exist.woff2"}}

	if err := inlineFonts(cfg); err == nil {
		t.Error("expected an error for a missing font file")
	}
}

func TestFetchFonts(t *testing.T) {
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/font.css", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `@font-face { font-family: 'Test'; font-weight: 400; src: url(%s/font.woff2) format('woff2'); }`, srv.URL)
	})
	mux.HandleFunc("/font.woff2", func(w http.ResponseWriter, r *http.Request) {
		w.Write(buildTestWoff2(t))
	})
	srv = httptest.NewServer(mux)
	defer srv.Close()

	t.Run("fetches configured font url and prints a faces block", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "cfg.yml")
		if err := os.WriteFile(path, []byte("justify: false\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		cfg.Font.URL = srv.URL + "/font.css"
		cfg.Font.Family = "Test"
		cfg.Code.FontURL = "" // isolate to the body font for this case

		var fetchErr error
		out := captureStdout(t, func() {
			fetchErr = fetchFonts(cfg, "fonts")
		})
		if fetchErr != nil {
			t.Fatalf("error: %v", fetchErr)
		}
		if !strings.Contains(out, "font:\n  faces:\n") {
			t.Errorf("expected a font faces block in output:\n%s", out)
		}
		if !strings.Contains(out, "weight: 400") {
			t.Errorf("expected weight in output:\n%s", out)
		}
		if !strings.Contains(out, "Add the block(s) above") {
			t.Errorf("expected the trailing instruction in output:\n%s", out)
		}

		entries, err := os.ReadDir(filepath.Join(dir, "fonts"))
		if err != nil {
			t.Fatalf("fonts dir not created: %v", err)
		}
		if len(entries) == 0 {
			t.Error("expected at least one font file written")
		}
	})

	t.Run("errors when nothing is configured to fetch", func(t *testing.T) {
		cfg := config.Default()
		cfg.Font.URL = ""
		cfg.Code.FontURL = ""

		if err := fetchFonts(cfg, "fonts"); err == nil {
			t.Error("expected an error when no font.url or code.font_url is set")
		}
	})

	t.Run("wraps the target key into the error", func(t *testing.T) {
		cfg := config.Default()
		cfg.Font.URL = "http://127.0.0.1:1/unreachable"
		cfg.Code.FontURL = ""

		err := fetchFonts(cfg, t.TempDir())
		if err == nil {
			t.Fatal("expected an error for an unreachable url")
		}
		if !strings.HasPrefix(err.Error(), "font: ") {
			t.Errorf("error = %q, want it prefixed with the target key", err.Error())
		}
	})

	t.Run("relative fonts dir resolves against the config directory", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "cfg.yml")
		if err := os.WriteFile(path, []byte("justify: false\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		cfg.Font.URL = srv.URL + "/font.css"
		cfg.Font.Family = "Test"
		cfg.Code.FontURL = ""

		captureStdout(t, func() {
			if err := fetchFonts(cfg, "./sub/fonts"); err != nil {
				t.Fatalf("error: %v", err)
			}
		})
		if _, err := os.Stat(filepath.Join(dir, "sub", "fonts")); err != nil {
			t.Errorf("expected fonts dir under the config directory: %v", err)
		}
	})
}

// buildTestWoff2 returns a minimal, structurally valid (but not renderable)
// WOFF2 byte buffer good enough for fontface.BuildCSS/FetchStylesheet, which
// only need a readable file with a recognized extension.
func buildTestWoff2(t *testing.T) []byte {
	t.Helper()
	header := make([]byte, 48)
	copy(header[0:4], "wOF2")
	header[13] = 1 // numTables = 1 (big-endian uint16 at offset 12-13)
	entry := []byte{0x3f, 'c', 'm', 'a', 'p', 0x0A}
	return append(header, entry...)
}
