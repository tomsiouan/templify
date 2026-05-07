package main

import (
	"log/slog"
	"os"

	"github.com/tomsiouan/templify/pkg/parser"
	"github.com/tomsiouan/templify/pkg/renderer"
	"github.com/tomsiouan/templify/pkg/tmpl"
)

func main() {
	cfg, err := parseFlags()
	if err != nil {
		slog.Error("invalid flags", "err", err)
		os.Exit(1)
	}

	setupLogger(cfg.Log)

	doc, err := parser.ParseFile(cfg.Input)
	if err != nil {
		slog.Error("parse failed", "err", err)
		os.Exit(1)
	}

	html, err := tmpl.Render(cfg.Template, doc)
	if err != nil {
		slog.Error("template failed", "err", err)
		os.Exit(1)
	}

	if err := renderer.ToPDF(html, cfg.Output); err != nil {
		slog.Error("render failed", "err", err)
		os.Exit(1)
	}

	slog.Info("generated", "output", cfg.Output)
}
