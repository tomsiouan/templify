package main

import (
	"flag"
	"fmt"
	"log/slog"
)

type LogConfig struct {
	Format string // "text" or "json"
	Level  slog.Level
}

type Config struct {
	Input    string
	Output   string
	Template string
	Mode     string
	TOC      bool
	Log      LogConfig
}

func parseFlags() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Input, "input", "", "path to the input Markdown file (required)")
	flag.StringVar(&cfg.Output, "output", "output.pdf", "path for the generated PDF")
	flag.StringVar(&cfg.Template, "template", "default", "built-in template name or path to a .html file")
	flag.StringVar(&cfg.Mode, "mode", "default", "rendering mode: default or report")
	var noTOC bool
	flag.BoolVar(&noTOC, "no-toc", false, "disable table of contents")
	flag.StringVar(&cfg.Log.Format, "log-format", "text", "log format: text or json")
	flag.TextVar(&cfg.Log.Level, "log-level", slog.LevelInfo, "log level: DEBUG, INFO, WARN, ERROR")
	flag.Parse()
	cfg.TOC = !noTOC

	if cfg.Input == "" {
		flag.Usage()
		return nil, fmt.Errorf("-input is required")
	}

	return &cfg, nil
}
