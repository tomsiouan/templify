package tmpl

import (
	"strings"
	"testing"

	"github.com/tomsiouan/templify/config"
	"github.com/tomsiouan/templify/document"
)

func TestRender(t *testing.T) {
	doc := &document.Document{Title: "Test Doc", Body: "<p>body</p>"}
	cfg := config.Default()

	t.Run("renders main template with document data", func(t *testing.T) {
		out, err := Render(`<title>{{.Title}}</title>`, "", doc, cfg)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if !strings.Contains(out, "Test Doc") {
			t.Errorf("expected title in output: %q", out)
		}
	})

	t.Run("invalid main template returns error", func(t *testing.T) {
		_, err := Render(`{{.Unclosed`, "", doc, cfg)
		if err == nil {
			t.Error("expected error for invalid template")
		}
	})

	t.Run("cover html injected into template data", func(t *testing.T) {
		cfg2 := config.Default()
		cfg2.Cover.Enabled = true
		out, err := Render(`{{.CoverHTML}}`, `<p>cover content</p>`, doc, cfg2)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if !strings.Contains(out, "cover content") {
			t.Errorf("expected cover content in output: %q", out)
		}
	})

	t.Run("config css injected into template data", func(t *testing.T) {
		out, err := Render(`{{.ConfigCSS}}`, "", doc, cfg)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if !strings.Contains(out, "<style>") {
			t.Errorf("expected style tag in ConfigCSS: %q", out)
		}
	})
}
