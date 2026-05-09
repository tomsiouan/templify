# templify

Convert Markdown files to styled PDFs using a config-driven HTML template system.

## Usage

```bash
templify -input document.md -output document.pdf
templify -input document.md -config config.yml -output document.pdf
templify -input document.md -template custom.html -output document.pdf
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-input` | — | Path to the input Markdown file (required) |
| `-output` | `output.pdf` | Path for the generated PDF |
| `-template` | `default` | Built-in template name or path to a custom `.html` template |
| `-config` | — | Path to a YAML config file |

## Configuration

All visual options are controlled via a YAML config file. See [`config.example.yml`](config.example.yml) for a full reference.

```yaml
page:
  size: A4
  margins:
    top: 25mm
    right: 20mm
    bottom: 25mm
    left: 25mm

font:
  family: Inter
  url: https://fonts.googleapis.com/css2?family=Inter:wght@400;600&display=swap
  size: 11pt
  line_height: 1.6

justify: false
paragraph_indent: 13mm   # first-line indent on paragraphs
heading_indent: 5mm      # padding-left step per sub-level (h3 = 1×, h4 = 2×, …)

heading_numbers:
  enabled: true           # prefix h2+ with 1 / 1.1 / 1.1.1 counters
  exclude:                # headings that must not be numbered
    - Introduction
    - Conclusion

toc:
  enabled: true
  max_depth: 3
  exclude:                # headings to omit from the TOC entirely
    - Remerciements

header:
  enabled: true
  background: "linear-gradient(to right, #1e293b, #475569)"
  # left/center/right set per-document in front matter

footer:
  enabled: true
  background: ""

colors:
  primary: "#1e293b"
  primary_light: "#475569"
  background: "#f8fafc"
  text: "#0f172a"
  text_muted: "#64748b"

cover:
  enabled: true
  template: ""            # path to a custom cover template (relative to this file)
```

## Front matter

Document metadata is declared as YAML at the top of the Markdown file. It drives the cover page and can override header/footer slots per document.

```markdown
---
title: My Report
author: Alice
date: 2026-05-06
header:
  left: "My Company"
  right: "Confidential"
footer:
  right: "{page} / {pages}"
---

## Introduction

Content starts here.
```

`{page}` and `{pages}` are replaced with the current page number and total page count.

## Markdown features

- **GFM** — tables, strikethrough, task lists
- **Footnotes** and **definition lists**
- **Typographer** — smart quotes, dashes
- **Image captions** — set the image title to render a `<figcaption>`:
  ```markdown
  ![alt text](image.png "This becomes the caption")
  ```
- **Auto heading IDs** — used for TOC anchor links

## Custom templates

Pass a `.html` file path to `-template` to use a fully custom template. Templates are Go `html/template` files with the following context:

```go
type Document struct {
    Title  string
    Author string
    Date   string
    Body   template.HTML    // rendered HTML body (do not re-escape)
    Meta   map[string]any   // all front matter fields
    TOC    []TocEntry
}

type TocEntry struct {
    Level    int
    Text     string
    ID       string
    NoNumber bool            // true when excluded from heading_numbers
}
```

The template also receives `.Config` (`*config.Config`) and `.ConfigCSS` (`template.HTML`) which injects the CSS derived from the config file.

## Installation

```bash
go install github.com/tomsiouan/templify@latest
```

Or build from source:

```bash
git clone https://github.com/tomsiouan/templify
cd templify
go build -o templify ./...
```

## Requirements

- Go 1.26.2+
- Chromium (auto-downloaded to `~/.cache/rod/` on first run)
