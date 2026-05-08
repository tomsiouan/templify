package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type HeadingLevel struct {
	Size            string `yaml:"size"`
	PageBreakBefore bool   `yaml:"page_break_before"`
}

type HeadingsConfig struct {
	H1 HeadingLevel `yaml:"h1"`
	H2 HeadingLevel `yaml:"h2"`
	H3 HeadingLevel `yaml:"h3"`
	H4 HeadingLevel `yaml:"h4"`
	H5 HeadingLevel `yaml:"h5"`
	H6 HeadingLevel `yaml:"h6"`
}

type MarginsConfig struct {
	Top    string `yaml:"top"`
	Right  string `yaml:"right"`
	Bottom string `yaml:"bottom"`
	Left   string `yaml:"left"`
}

type PageConfig struct {
	Size    string        `yaml:"size"`
	Margins MarginsConfig `yaml:"margins"`
}

type FontConfig struct {
	Family     string  `yaml:"family"`
	URL        string  `yaml:"url"`
	Size       string  `yaml:"size"`
	LineHeight float64 `yaml:"line_height"`
}

type TOCConfig struct {
	Enabled  bool `yaml:"enabled"`
	MaxDepth int  `yaml:"max_depth"`
}

// HeaderFooterConfig configures a header or footer band on every page.
// Left/Center/Right accept plain text or {page}/{pages} tokens.
type HeaderFooterConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Background string `yaml:"background"`
	Left       string `yaml:"left"`
	Center     string `yaml:"center"`
	Right      string `yaml:"right"`
}

type ColorsConfig struct {
	Primary      string `yaml:"primary"`
	PrimaryLight string `yaml:"primary_light"`
	Background   string `yaml:"background"`
	Text         string `yaml:"text"`
	TextMuted    string `yaml:"text_muted"`
}

type CoverConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Template string `yaml:"template"`
}

type Config struct {
	Page      PageConfig         `yaml:"page"`
	Font      FontConfig         `yaml:"font"`
	Headings  HeadingsConfig     `yaml:"headings"`
	Justify   bool               `yaml:"justify"`
	TOC       TOCConfig          `yaml:"toc"`
	Header    HeaderFooterConfig `yaml:"header"`
	Footer    HeaderFooterConfig `yaml:"footer"`
	BlankPage bool               `yaml:"blank_page"`
	Colors    ColorsConfig       `yaml:"colors"`
	Cover     CoverConfig        `yaml:"cover"`
	dir       string
}

// ApplyDocMeta overrides header/footer Left/Center/Right from the document's front matter.
// Front matter keys "header" and "footer" accept the same left/center/right sub-keys.
func (c *Config) ApplyDocMeta(meta map[string]any) {
	if v, ok := meta["header"]; ok {
		if m := toStringMap(v); m != nil {
			if s, ok := m["left"].(string); ok {
				c.Header.Left = s
			}
			if s, ok := m["center"].(string); ok {
				c.Header.Center = s
			}
			if s, ok := m["right"].(string); ok {
				c.Header.Right = s
			}
		}
	}
	if v, ok := meta["footer"]; ok {
		if m := toStringMap(v); m != nil {
			if s, ok := m["left"].(string); ok {
				c.Footer.Left = s
			}
			if s, ok := m["center"].(string); ok {
				c.Footer.Center = s
			}
			if s, ok := m["right"].(string); ok {
				c.Footer.Right = s
			}
		}
	}
}

// toStringMap converts map[interface{}]interface{} (yaml.v2) or map[string]any to map[string]any.
func toStringMap(v any) map[string]any {
	switch m := v.(type) {
	case map[string]any:
		return m
	case map[interface{}]interface{}:
		out := make(map[string]any, len(m))
		for k, v := range m {
			if ks, ok := k.(string); ok {
				out[ks] = v
			}
		}
		return out
	}
	return nil
}

// ResolvePath resolves a path relative to the config file's directory.
func (c *Config) ResolvePath(p string) string {
	if filepath.IsAbs(p) || c.dir == "" {
		return p
	}
	return filepath.Join(c.dir, p)
}

func Default() *Config {
	return &Config{
		Page: PageConfig{
			Size: "A4",
			Margins: MarginsConfig{
				Top:    "25mm",
				Right:  "20mm",
				Bottom: "25mm",
				Left:   "25mm",
			},
		},
		Font: FontConfig{
			Family:     "Inter",
			URL:        "https://fonts.googleapis.com/css2?family=Inter:wght@400;600&display=swap",
			Size:       "11pt",
			LineHeight: 1.6,
		},
		Headings: HeadingsConfig{
			H1: HeadingLevel{Size: "24pt"},
			H2: HeadingLevel{Size: "18pt"},
			H3: HeadingLevel{Size: "14pt"},
			H4: HeadingLevel{Size: "12pt"},
			H5: HeadingLevel{Size: "11pt"},
			H6: HeadingLevel{Size: "10pt"},
		},
		Justify: false,
		TOC: TOCConfig{
			Enabled:  true,
			MaxDepth: 2,
		},
		Header: HeaderFooterConfig{
			Enabled: false,
		},
		Footer: HeaderFooterConfig{
			Enabled: true,
			Right:   "{page} / {pages}",
		},
		BlankPage: false,
		Colors: ColorsConfig{
			Primary:      "#1e293b",
			PrimaryLight: "#475569",
			Background:   "#f8fafc",
			Text:         "#0f172a",
			TextMuted:    "#64748b",
		},
		Cover: CoverConfig{
			Enabled: true,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}
	cfg.dir = filepath.Dir(path)
	return cfg, nil
}

// ContentCSS converts a header/footer slot value to a CSS content value.
// Empty string produces `""`. Tokens {page} and {pages} become CSS counters.
func ContentCSS(s string) string {
	if s == "" {
		return `""`
	}
	re := regexp.MustCompile(`\{page\}|\{pages\}`)
	var sb strings.Builder
	last := 0
	for _, m := range re.FindAllStringIndex(s, -1) {
		if m[0] > last {
			sb.WriteString(`"`)
			sb.WriteString(s[last:m[0]])
			sb.WriteString(`" `)
		}
		if s[m[0]:m[1]] == "{page}" {
			sb.WriteString("counter(page) ")
		} else {
			sb.WriteString("counter(pages) ")
		}
		last = m[1]
	}
	if last < len(s) {
		sb.WriteString(`"`)
		sb.WriteString(s[last:])
		sb.WriteString(`"`)
	}
	return strings.TrimSpace(sb.String())
}
