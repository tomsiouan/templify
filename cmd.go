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
	Input      string
	Output     string
	Template   string
	ConfigPath string
	Log        LogConfig
}

func parseFlags() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Input, "input", "", "path to the input Markdown file (required)")
	flag.StringVar(&cfg.Output, "output", "output.pdf", "path for the generated PDF")
	flag.StringVar(&cfg.Template, "template", "default", "built-in template name or path to a .html file")
	flag.StringVar(&cfg.ConfigPath, "config", "", "path to a YAML config file")
	flag.StringVar(&cfg.Log.Format, "log-format", "text", "log format: text or json")
	flag.TextVar(&cfg.Log.Level, "log-level", slog.LevelInfo, "log level: DEBUG, INFO, WARN, ERROR")
	flag.Parse()

	if cfg.Input == "" {
		flag.Usage()
		return nil, fmt.Errorf("-input is required")
	}

	return &cfg, nil
}
