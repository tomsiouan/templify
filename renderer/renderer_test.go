package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectPagedJS(t *testing.T) {
	t.Run("injects script before closing head tag", func(t *testing.T) {
		html := "<html><head></head><body></body></html>"
		result := injectPagedJS(html, []byte("console.log('paged')"))
		if !strings.Contains(result, "console.log('paged')") {
			t.Errorf("paged.js content missing: %q", result)
		}
		if !strings.Contains(result, "</head>") {
			t.Errorf("</head> should still be present: %q", result)
		}
		scriptIdx := strings.Index(result, "<script>")
		headIdx := strings.Index(result, "</head>")
		if scriptIdx > headIdx {
			t.Errorf("script should appear before </head>")
		}
	})

	t.Run("no head tag leaves html unchanged", func(t *testing.T) {
		html := "<html><body>no head</body></html>"
		result := injectPagedJS(html, []byte("paged"))
		if result != html {
			t.Errorf("expected unchanged html, got %q", result)
		}
	})

	t.Run("injects pagedJSConfig alongside polyfill", func(t *testing.T) {
		html := "<html><head></head></html>"
		result := injectPagedJS(html, []byte("polyfill"))
		if !strings.Contains(result, "PagedConfig") {
			t.Errorf("expected PagedConfig in result: %q", result)
		}
	})
}

func TestWriteTempHTML(t *testing.T) {
	t.Run("creates temp file with html content", func(t *testing.T) {
		dir := t.TempDir()
		html := "<html><body>test</body></html>"
		path, cleanup, err := writeTempHTML(html, dir)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		defer cleanup()

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("could not read temp file: %v", err)
		}
		if string(data) != html {
			t.Errorf("content = %q, want %q", string(data), html)
		}
	})

	t.Run("temp file is in baseDir", func(t *testing.T) {
		dir := t.TempDir()
		path, cleanup, err := writeTempHTML("content", dir)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
		if filepath.Dir(path) != dir {
			t.Errorf("expected file in %q, got %q", dir, filepath.Dir(path))
		}
	})

	t.Run("cleanup removes the temp file", func(t *testing.T) {
		dir := t.TempDir()
		path, cleanup, err := writeTempHTML("content", dir)
		if err != nil {
			t.Fatal(err)
		}
		cleanup()
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("temp file should be removed after cleanup")
		}
	})
}
