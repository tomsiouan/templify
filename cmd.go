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
	Bundle     string
	ConfigPath string
	Log        LogConfig
}

func parseFlags() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Input, "input", "", "path to the input Markdown file (required)")
	flag.StringVar(&cfg.Output, "output", "output.pdf", "path for the generated PDF")
	flag.StringVar(&cfg.Bundle, "bundle", "report", "built-in bundle name or path to a bundle directory")
	flag.StringVar(&cfg.ConfigPath, "config", "", "path to a YAML config file (overrides bundle defaults)")
	flag.StringVar(&cfg.Log.Format, "log-format", "text", "log format: text or json")
	flag.TextVar(&cfg.Log.Level, "log-level", slog.LevelInfo, "log level: DEBUG, INFO, WARN, ERROR")
	flag.Parse()

	if cfg.Input == "" {
		flag.Usage()
		return nil, fmt.Errorf("-input is required")
	}

	return &cfg, nil
}
