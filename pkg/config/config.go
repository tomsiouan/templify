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
	Enabled  bool     `yaml:"enabled"`
	MaxDepth int      `yaml:"max_depth"`
	Exclude  []string `yaml:"exclude"`
	PreTOC   []string `yaml:"pre_toc"` // sections extracted from body and placed before the TOC
}

type HeadingNumbersConfig struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
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

type ReferencesConfig struct {
	Bibliography string `yaml:"bibliography"` // h2 heading text of the bibliography section
	Sitography   string `yaml:"sitography"`   // h2 heading text of the sitography section
	Figures      string `yaml:"figures"`      // h2 heading text for the auto-generated table of figures
}

type Config struct {
	Page            PageConfig         `yaml:"page"`
	Font            FontConfig         `yaml:"font"`
	Headings        HeadingsConfig     `yaml:"headings"`
	Justify         bool                 `yaml:"justify"`
	ParagraphIndent string               `yaml:"paragraph_indent"`
	HeadingIndent   string               `yaml:"heading_indent"`
	HeadingNumbers  HeadingNumbersConfig `yaml:"heading_numbers"`
	TOC             TOCConfig          `yaml:"toc"`
	Header          HeaderFooterConfig `yaml:"header"`
	Footer          HeaderFooterConfig `yaml:"footer"`
	BlankPage       bool               `yaml:"blank_page"`
	Colors          ColorsConfig       `yaml:"colors"`
	Cover           CoverConfig        `yaml:"cover"`
	References      ReferencesConfig   `yaml:"references"`
	Custom          map[string]any     `yaml:"custom"`
	dir             string
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

// CustomString returns a string value from Config.Custom at the given dot-separated path,
// falling back to def if missing or not a string (e.g. "invoice.labels.client").
func (c *Config) CustomString(path, def string) string {
	parts := strings.SplitN(path, ".", 2)
	if c.Custom == nil {
		return def
	}
	v, ok := c.Custom[parts[0]]
	if !ok {
		return def
	}
	if len(parts) == 1 {
		if s, ok := v.(string); ok {
			return s
		}
		return def
	}
	sub := toStringMap(v)
	if sub == nil {
		return def
	}
	nested := &Config{Custom: sub}
	return nested.CustomString(parts[1], def)
}

func (c *Config) CustomBool(path string, def bool) bool {
	parts := strings.SplitN(path, ".", 2)
	if c.Custom == nil {
		return def
	}
	v, ok := c.Custom[parts[0]]
	if !ok {
		return def
	}
	if len(parts) == 1 {
		if b, ok := v.(bool); ok {
			return b
		}
		return def
	}
	sub := toStringMap(v)
	if sub == nil {
		return def
	}
	nested := &Config{Custom: sub}
	return nested.CustomBool(parts[1], def)
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
	if err := mergeYAML(cfg, path); err != nil {
		return nil, err
	}
	cfg.dir = filepath.Dir(path)
	return cfg, nil
}

// FromBundle builds a Config from the raw YAML of a bundle's default.yml.
// Falls back to Default() if data is nil.
func FromBundle(data []byte) (*Config, error) {
	cfg := Default()
	if len(data) == 0 {
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse bundle config: %w", err)
	}
	return cfg, nil
}

// MergeFile overlays the YAML file at path on top of cfg.
func MergeFile(cfg *Config, path string) error {
	if err := mergeYAML(cfg, path); err != nil {
		return err
	}
	cfg.dir = filepath.Dir(path)
	return nil
}

func mergeYAML(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config %q: %w", path, err)
	}
	return nil
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
