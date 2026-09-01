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
	Input          string
	Output         string
	Bundle         string
	ConfigPath     string
	ListCodeThemes bool
	FetchFonts     bool
	FontsDir       string
	ShowVersion    bool
	Log            LogConfig
}

func parseFlags() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Input, "input", "", "path to the input Markdown file (required)")
	flag.StringVar(&cfg.Output, "output", "output.pdf", "path for the generated PDF")
	flag.StringVar(&cfg.Bundle, "bundle", "report", "built-in bundle name or path to a bundle directory")
	flag.StringVar(&cfg.ConfigPath, "config", "", "path to a YAML config file (overrides bundle defaults)")
	flag.BoolVar(&cfg.ListCodeThemes, "code-themes", false, "list the syntax highlighting themes available for code.theme and exit")
	flag.BoolVar(&cfg.FetchFonts, "fetch-fonts", false, "download the fonts behind font.url/code.font_url, print the matching faces: block, and exit")
	flag.StringVar(&cfg.FontsDir, "fonts-dir", "./fonts", "directory to save fonts into, for -fetch-fonts (relative to the config file)")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "print the version and exit")
	flag.StringVar(&cfg.Log.Format, "log-format", "text", "log format: text or json")
	flag.TextVar(&cfg.Log.Level, "log-level", slog.LevelInfo, "log level: DEBUG, INFO, WARN, ERROR")
	flag.Parse()

	if cfg.ShowVersion {
		return &cfg, nil
	}

	if cfg.ListCodeThemes {
		return &cfg, nil
	}

	if cfg.FetchFonts {
		if cfg.ConfigPath == "" {
			return nil, fmt.Errorf("-fetch-fonts requires -config")
		}
		return &cfg, nil
	}

	if cfg.Input == "" {
		flag.Usage()
		return nil, fmt.Errorf("-input is required")
	}

	return &cfg, nil
}
