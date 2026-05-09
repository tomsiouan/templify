package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/tomsiouan/templify/pkg/config"
	"github.com/tomsiouan/templify/pkg/parser"
	"github.com/tomsiouan/templify/pkg/renderer"
	"github.com/tomsiouan/templify/pkg/tmpl"
)

func main() {
	flags, err := parseFlags()
	if err != nil {
		slog.Error("invalid flags", "err", err)
		os.Exit(1)
	}

	setupLogger(flags.Log)

	var cfg *config.Config
	if flags.ConfigPath != "" {
		cfg, err = config.Load(flags.ConfigPath)
		if err != nil {
			slog.Error("load config failed", "err", err)
			os.Exit(1)
		}
	} else {
		cfg = config.Default()
	}

	doc, err := parser.ParseFile(flags.Input)
	if err != nil {
		slog.Error("parse failed", "err", err)
		os.Exit(1)
	}

	cfg.ApplyDocMeta(doc.Meta)

	if cfg.HeadingNumbers.Enabled {
		parser.MarkNoNumber(doc, cfg.HeadingNumbers.Exclude)
	}

	if cfg.TOC.Enabled {
		tocExclude := make(map[string]bool, len(cfg.TOC.Exclude))
		for _, s := range cfg.TOC.Exclude {
			tocExclude[s] = true
		}
		filtered := doc.TOC[:0]
		for _, e := range doc.TOC {
			if e.Level <= cfg.TOC.MaxDepth && !tocExclude[e.Text] {
				filtered = append(filtered, e)
			}
		}
		doc.TOC = filtered
	} else {
		doc.TOC = nil
	}

	parser.ExtractPreTOC(doc, cfg.TOC.PreTOC)
	parser.InlineFootnotes(doc)

	html, err := tmpl.Render(flags.Template, doc, cfg)
	if err != nil {
		slog.Error("template failed", "err", err)
		os.Exit(1)
	}

	if err := renderer.ToPDF(html, flags.Output, filepath.Dir(flags.Input)); err != nil {
		slog.Error("render failed", "err", err)
		os.Exit(1)
	}

	slog.Info("generated", "output", flags.Output)
}
