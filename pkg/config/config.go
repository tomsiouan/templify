package config

import (
	"fmt"
	"os"
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

type PageNumbersConfig struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"`
}

type ColorsConfig struct {
	Primary      string `yaml:"primary"`
	PrimaryLight string `yaml:"primary_light"`
	Background   string `yaml:"background"`
	Text         string `yaml:"text"`
	TextMuted    string `yaml:"text_muted"`
}

type Config struct {
	Page        PageConfig        `yaml:"page"`
	Font        FontConfig        `yaml:"font"`
	Headings    HeadingsConfig    `yaml:"headings"`
	Justify     bool              `yaml:"justify"`
	TOC         TOCConfig         `yaml:"toc"`
	PageNumbers PageNumbersConfig `yaml:"page_numbers"`
	BlankPage   bool              `yaml:"blank_page"`
	Colors      ColorsConfig      `yaml:"colors"`
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
		PageNumbers: PageNumbersConfig{
			Enabled: true,
			Format:  "{page} / {pages}",
		},
		BlankPage: false,
		Colors: ColorsConfig{
			Primary:      "#1e293b",
			PrimaryLight: "#475569",
			Background:   "#f8fafc",
			Text:         "#0f172a",
			TextMuted:    "#64748b",
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
	return cfg, nil
}

// PageNumberCSS converts a format string like "{page} / {pages}" to a CSS content value
// e.g. counter(page) " / " counter(pages)
func PageNumberCSS(format string) string {
	re := regexp.MustCompile(`\{page\}|\{pages\}`)
	var sb strings.Builder
	last := 0
	for _, m := range re.FindAllStringIndex(format, -1) {
		if m[0] > last {
			sb.WriteString(`"`)
			sb.WriteString(format[last:m[0]])
			sb.WriteString(`" `)
		}
		if format[m[0]:m[1]] == "{page}" {
			sb.WriteString("counter(page) ")
		} else {
			sb.WriteString("counter(pages) ")
		}
		last = m[1]
	}
	if last < len(format) {
		sb.WriteString(`"`)
		sb.WriteString(format[last:])
		sb.WriteString(`"`)
	}
	return strings.TrimSpace(sb.String())
}
