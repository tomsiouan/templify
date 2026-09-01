package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/tomsiouan/templify/config"
	"github.com/tomsiouan/templify/fontface"
)

// inlineFonts embeds the configured self-hosted faces into the document
// stylesheet and drops the matching webfont URLs, so a family is never loaded
// twice. A linked stylesheet leaves the PDF text layer split glyph by glyph;
// the inlined faces are what keep copied text readable.
func inlineFonts(cfg *config.Config) error {
	css, err := fontface.BuildCSS(cfg)
	if err != nil {
		return err
	}
	if css == "" {
		if cfg.Font.URL != "" || cfg.Code.FontURL != "" {
			slog.Debug("fonts loaded from a stylesheet; copied text may come out garbled",
				"hint", "run templify -fetch-fonts to self-host them")
		}
		return nil
	}
	cfg.SetFontFacesCSS(css)

	// Bundles emit the <link> from Config.Font.URL; clearing it is how the
	// linked copy is suppressed without every bundle needing to know about
	// faces. The code font's @import is suppressed in the CSS generator.
	if len(cfg.Font.Faces) > 0 {
		cfg.Font.URL = ""
	}
	return nil
}

// fetchFonts downloads the font files behind the configured webfont stylesheets
// into dir and prints the `faces:` blocks to add to the config.
//
// The config is deliberately left untouched: yaml.v2 cannot round-trip
// comments, and these files are usually commented.
func fetchFonts(cfg *config.Config, dir string) error {
	targets := []struct {
		key    string
		url    string
		family string
	}{
		{"font", cfg.Font.URL, cfg.Font.Family},
		{"code", cfg.Code.FontURL, cfg.Code.FontFamily},
	}

	base := cfg.Dir()
	if base == "" {
		base = "."
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(base, dir)
	}

	found := false
	for _, t := range targets {
		if t.url == "" {
			continue
		}
		found = true
		slog.Info("fetching", "font", t.family, "url", t.url)

		fetched, err := fontface.FetchStylesheet(t.url, dir, fontface.Prefix(t.family))
		if err != nil {
			return fmt.Errorf("%s: %w", t.key, err)
		}
		slog.Info("downloaded", "font", t.family, "files", len(fetched), "dir", dir)

		fmt.Println()
		fmt.Print(fontface.FacesYAML(t.key, fetched, base))
	}

	if !found {
		return fmt.Errorf("no font.url or code.font_url to fetch")
	}

	fmt.Println()
	fmt.Println("# Add the block(s) above to your config, and drop the matching url: line.")
	return nil
}
