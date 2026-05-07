package main

import (
	"log/slog"
	"os"
	"strings"
)

func setupLogger(cfg LogConfig) {
	opts := &slog.HandlerOptions{Level: cfg.Level}
	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler))
}
