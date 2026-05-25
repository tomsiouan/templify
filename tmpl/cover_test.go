package tmpl

import (
	"strings"
	"testing"

	"github.com/tomsiouan/templify/config"
	"github.com/tomsiouan/templify/document"
)

func TestRenderCover(t *testing.T) {
	doc := &document.Document{Title: "My Doc"}

	t.Run("disabled cover returns empty", func(t *testing.T) {
		cfg := config.Default()
		cfg.Cover.Enabled = false
		html, css, err := renderCover("<p>cover</p>", doc, cfg)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if html != "" || css != "" {
			t.Errorf("expected empty html and css, got html=%q css=%q", html, css)
		}
	})

	t.Run("empty content returns empty", func(t *testing.T) {
		cfg := config.Default()
		cfg.Cover.Enabled = true
		html, css, err := renderCover("", doc, cfg)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if html != "" || css != "" {
			t.Errorf("expected empty html and css")
		}
	})

	t.Run("valid template renders html", func(t *testing.T) {
		cfg := config.Default()
		cfg.Cover.Enabled = true
		html, _, err := renderCover(`<h1>{{.Title}}</h1>`, doc, cfg)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if !strings.Contains(string(html), "My Doc") {
			t.Errorf("expected title in rendered html: %q", html)
		}
	})

	t.Run("style tags extracted into css", func(t *testing.T) {
		cfg := config.Default()
		cfg.Cover.Enabled = true
		content := `<style>body { color: red; }</style><p>text</p>`
		html, css, err := renderCover(content, doc, cfg)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if strings.Contains(string(html), "<style>") {
			t.Errorf("style tag should be removed from html: %q", html)
		}
		if !strings.Contains(string(css), "color: red") {
			t.Errorf("style content should be in css: %q", css)
		}
	})

	t.Run("invalid template returns error", func(t *testing.T) {
		cfg := config.Default()
		cfg.Cover.Enabled = true
		_, _, err := renderCover(`{{.Unclosed`, doc, cfg)
		if err == nil {
			t.Error("expected error for invalid template")
		}
	})
}
