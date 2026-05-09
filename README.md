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

All visual options are controlled via a YAML config file.

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
  pre_toc:                # sections extracted from body and placed before the TOC
    - Remerciements

references:
  bibliography: "Bibliographie"       # h2 heading of the bibliography section
  sitography: "Sitographie"           # h2 heading of the sitography section
  figures: "Table des illustrations"  # h2 heading to replace with the auto-generated table of figures

header:
  enabled: true
  background: "linear-gradient(to right, #1e293b, #475569)"
  # left/center/right set per-document in front matter

footer:
  enabled: true
  background: ""

blank_page: true          # insert a blank page before the TOC (double-sided printing)

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
- **Footnotes** — rendered at the bottom of the page where the reference appears
- **Definition lists**
- **Typographer** — smart quotes, dashes
- **Image captions** — set the image `title` attribute to render a `<figcaption>`:
  ```markdown
  ![alt text](image.png "This becomes the caption")
  ```
- **Auto heading IDs** — used for TOC anchor links

## Back matter

### Table of figures

Set `references.figures` to the h2 heading text in your document. The tool replaces the section's content with an auto-generated table (dotted leaders, page numbers resolved at render time):

```markdown
## Table des illustrations

<!-- content is replaced automatically -->
```

### Bibliography & sitography

Set `references.bibliography` and/or `references.sitography` to the corresponding h2 heading texts. List items are formatted as `[n]` numbered references.

Use **ordered lists** so the numbers are visible in your source — you can then reference `[3]` in the text without guessing:

```markdown
## Bibliographie

1. CommonMark Spec — [spec.commonmark.org](https://spec.commonmark.org)
2. CSS Paged Media Module Level 3 — W3C Working Draft

## Sitographie

3. goldmark : [github.com/yuin/goldmark](https://github.com/yuin/goldmark)
4. paged.js : [pagedjs.org](https://pagedjs.org)
```

For sitography, any `http(s)` link in a list item is automatically moved to a new line below the entry title. Optional h3 sub-sections are supported — numbering is continuous across all sub-sections.

## Custom templates

Pass a `.html` file path to `-template` to use a fully custom template. Templates are Go `html/template` files with the following context:

```go
type Document struct {
    Title   string
    Author  string
    Date    string
    Body    template.HTML    // rendered HTML body (do not re-escape)
    PreTOC  template.HTML    // pre-TOC sections extracted from body
    Meta    map[string]any   // all front matter fields
    TOC     []TocEntry
    Figures []FigureEntry
}

type TocEntry struct {
    Level    int
    Text     string
    ID       string
    NoNumber bool            // true when excluded from heading_numbers
}

type FigureEntry struct {
    ID      string           // e.g. "fig-1"
    Caption string
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
