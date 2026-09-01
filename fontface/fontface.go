// Package fontface inlines self-hosted font files into the document stylesheet
// and fetches the files a webfont stylesheet points at.
//
// Fonts loaded through a linked stylesheet make Chromium emit one positioned
// run per glyph in the PDF, which readers reassemble out of order: copied text
// comes back with letters transposed and stray spaces. The same font declared
// as an inline @font-face with the file embedded as a data URI produces a
// contiguous, correctly ordered text layer.
package fontface

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomsiouan/templify/config"
)

// mimeTypes maps a font file extension to the MIME type used in its data URI.
var mimeTypes = map[string]string{
	".woff2": "font/woff2",
	".woff":  "font/woff",
	".ttf":   "font/ttf",
	".otf":   "font/otf",
}

// formatNames maps a font file extension to the CSS format() hint, which tells
// the browser what it is being handed without sniffing the bytes.
var formatNames = map[string]string{
	".woff2": "woff2",
	".woff":  "woff",
	".ttf":   "truetype",
	".otf":   "opentype",
}

// BuildCSS returns the @font-face rules for every face configured on the body
// and code fonts, with each file read and embedded as a data URI. It returns ""
// when no face is configured. Paths are resolved relative to the config file.
func BuildCSS(cfg *config.Config) (string, error) {
	var sb strings.Builder

	if err := writeFaces(&sb, cfg, cfg.Font.Family, cfg.Font.Faces); err != nil {
		return "", fmt.Errorf("body font: %w", err)
	}
	if err := writeFaces(&sb, cfg, cfg.Code.FontFamily, cfg.Code.Faces); err != nil {
		return "", fmt.Errorf("code font: %w", err)
	}

	return sb.String(), nil
}

// writeFaces appends one @font-face rule per face, all for the same family.
func writeFaces(sb *strings.Builder, cfg *config.Config, family string, faces []config.FontFace) error {
	family = strings.TrimSpace(family)
	if len(faces) == 0 {
		return nil
	}
	if family == "" {
		return fmt.Errorf("faces are configured but the family name is empty")
	}

	for _, face := range faces {
		if strings.TrimSpace(face.File) == "" {
			return fmt.Errorf("face with no file")
		}
		path := cfg.ResolvePath(face.File)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %q: %w", path, err)
		}

		ext := strings.ToLower(filepath.Ext(path))
		mime, ok := mimeTypes[ext]
		if !ok {
			return fmt.Errorf("%q: unsupported font format %q, want one of .woff2 .woff .ttf .otf", path, ext)
		}

		weight := face.Weight
		if weight == 0 {
			weight = 400
		}
		style := strings.TrimSpace(face.Style)
		if style == "" {
			style = "normal"
		}

		fmt.Fprintf(sb, "@font-face {\n")
		fmt.Fprintf(sb, "    font-family: %q;\n", family)
		fmt.Fprintf(sb, "    font-style: %s;\n", style)
		fmt.Fprintf(sb, "    font-weight: %d;\n", weight)
		fmt.Fprintf(sb, "    font-display: block;\n")
		fmt.Fprintf(sb, "    src: url(data:%s;base64,%s) format(%q);\n",
			mime, base64.StdEncoding.EncodeToString(data), formatNames[ext])
		if r := strings.TrimSpace(face.UnicodeRange); r != "" {
			fmt.Fprintf(sb, "    unicode-range: %s;\n", r)
		}
		fmt.Fprintf(sb, "}\n\n")
	}
	return nil
}
