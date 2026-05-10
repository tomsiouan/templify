---
title: Getting Started
weight: 1
---

## Installation

```bash
go install github.com/tomsiouan/templify@latest
```

Or build from source:

```bash
git clone https://github.com/tomsiouan/templify
cd templify
go build -o build/templify .
```

Chromium is auto-downloaded to `~/.cache/rod/` on first run. No other dependencies required.

## Requirements

- Go 1.26.2+

## Your first PDF

```bash
templify -input document.md -output document.pdf
```

With a config file:

```bash
templify -input document.md -bundle report -config config.yml -output document.pdf
```

## CLI flags

| Flag | Default | Description |
|---|---|---|
| `-input` | — | Path to the input Markdown file (required) |
| `-output` | `output.pdf` | Path for the generated PDF |
| `-bundle` | `report` | Built-in bundle name or path to a local bundle directory |
| `-config` | — | YAML config file overlaid on top of the bundle defaults |

## Built-in bundles

| Bundle | Description |
|---|---|
| `report` | Multi-page document with cover, TOC, and header/footer |
| `invoice` | Single-page invoice with automatic HT/TVA/TTC totals |
| `quote` | Single-page quote with optional VAT display |
