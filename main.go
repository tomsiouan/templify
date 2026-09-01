package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/tomsiouan/templify/bundle"
	"github.com/tomsiouan/templify/config"
	"github.com/tomsiouan/templify/highlight"
	"github.com/tomsiouan/templify/parser"
	"github.com/tomsiouan/templify/renderer"
	"github.com/tomsiouan/templify/tmpl"
)

func main() {
	flags, err := parseFlags()
	if err != nil {
		slog.Error("invalid flags", "err", err)
		os.Exit(1)
	}

	setupLogger(flags.Log)

	if flags.ListCodeThemes {
		for _, name := range highlight.Themes() {
			fmt.Println(name)
		}
		return
	}

	b, err := bundle.Load(flags.Bundle)
	if err != nil {
		slog.Error("load bundle failed", "err", err)
		os.Exit(1)
	}

	// Start from bundle defaults, then apply user config on top.
	cfg, err := config.FromBundle(b.Config)
	if err != nil {
		slog.Error("load bundle config failed", "err", err)
		os.Exit(1)
	}
	if flags.ConfigPath != "" {
		if err := config.MergeFile(cfg, flags.ConfigPath); err != nil {
			slog.Error("load config failed", "err", err)
			os.Exit(1)
		}
	}

	if flags.FetchFonts {
		if err := fetchFonts(cfg, flags.FontsDir); err != nil {
			slog.Error("fetch fonts failed", "err", err)
			os.Exit(1)
		}
		return
	}

	if err := inlineFonts(cfg); err != nil {
		slog.Error("inline fonts failed", "err", err)
		os.Exit(1)
	}

	// Override cover template if the bundle ships one and user hasn't set a custom path.
	if b.Cover != "" && cfg.Cover.Template == "" {
		cfg.Cover.Enabled = true
	}

	doc, err := parser.ParseFile(flags.Input, cfg.Code)
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

	// Before ExtractPreTOC, so code blocks moved out of the body are tagged too.
	parser.MarkShortCodeBlocks(doc, cfg.Code.KeepTogetherLines)

	parser.ExtractPreTOC(doc, cfg.TOC.PreTOC)
	parser.InlineFootnotes(doc)
	parser.ProcessFigures(doc)
	parser.GenerateFigureTable(doc, cfg.References.Figures)
	parser.NumberReferences(doc, cfg.References.Bibliography, cfg.References.Sitography)

	coverContent := b.Cover
	if cfg.Cover.Template != "" {
		data, err := os.ReadFile(cfg.ResolvePath(cfg.Cover.Template))
		if err != nil {
			slog.Error("read cover template failed", "err", err)
			os.Exit(1)
		}
		coverContent = string(data)
	}

	html, err := tmpl.Render(b.Main, coverContent, doc, cfg)
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
