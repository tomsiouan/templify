package main

import (
	"flag"
	"fmt"
)

type LogConfig struct {
	Format string // "text" or "json"
	Level  string // "debug", "info", "warn", "error"
}

type Config struct {
	Input    string
	Output   string
	Template string
	Log      LogConfig
}

func parseFlags() (*Config, error) {
	input := flag.String("input", "", "path to the input Markdown file (required)")
	output := flag.String("output", "output.pdf", "path for the generated PDF")
	template := flag.String("template", "default", "built-in template name or path to a .html file")
	logFormat := flag.String("log-format", "text", "log format: text or json")
	logLevel := flag.String("log-level", "info", "log level: debug, info, warn, error")
	flag.Parse()

	if *input == "" {
		flag.Usage()
		return nil, fmt.Errorf("-input is required")
	}

	return &Config{
		Input:    *input,
		Output:   *output,
		Template: *template,
		Log:      LogConfig{Format: *logFormat, Level: *logLevel},
	}, nil
}
