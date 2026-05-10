# templify

Convert Markdown files to styled PDFs using a bundle-driven template system.

## Usage

```bash
templify -input document.md -output document.pdf
templify -input document.md -bundle report -config config.yml -output document.pdf
templify -input invoice.md -bundle invoice -output invoice.pdf
templify -input quote.md -bundle quote -output quote.pdf
templify -input document.md -bundle ./my-bundle/ -output document.pdf
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-input` | — | Path to the input Markdown file (required) |
| `-output` | `output.pdf` | Path for the generated PDF |
| `-bundle` | `report` | Built-in bundle name (`report`, `invoice`, `quote`) or path to a local bundle directory |
| `-config` | — | Path to a YAML config file overlaid on top of the bundle defaults |

## Bundles

A bundle is a directory containing a `main.html` template, an optional `cover.html`, and an optional `default.yml` with bundle-specific config defaults. Built-in bundles are embedded in the binary.

| Bundle | Description |
|---|---|
| `report` | Multi-page document with cover page, TOC, and header/footer |
| `invoice` | Single-page invoice with automatic HT/TVA/TTC calculation |
| `quote` | Single-page quote (devis) with optional VAT display |

### Custom bundle

Create a directory with at least a `main.html` and pass its path as `-bundle`:

```
my-bundle/
├── main.html       # Go html/template
├── cover.html      # optional cover page
└── default.yml     # optional config defaults
```

```bash
templify -input document.md -bundle ./my-bundle/ -output document.pdf
```

## Configuration

All visual options are controlled via a YAML config file. Values are layered: `config.Default()` → bundle `default.yml` → user `-config` file → front matter.

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

custom:                   # bundle-specific options, accessible in templates via .Config.CustomString / .Config.CustomBool
  invoice:
    show:
      logo: true
    logo: ./logo.png
    labels:
      client: "Bill to"
```

## Front matter

Document metadata is declared as YAML at the top of the Markdown file.

### Report

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

### Invoice

```markdown
---
invoice_number: "2026-001"
date: 2026-05-10
due_date: 2026-06-09
vat: 20
---

## Vendeur

**Acme SAS**
12 rue de la Paix — 75001 Paris

## Client

**Société Exemple**
45 avenue des Champs — 69000 Lyon

## Articles

| Description | Qté | Prix HT |
|---|---|---|
| Développement API | 10 | 950,00 |
| Formation (demi-journée) | 0,5 | 800,00 |

## Conditions

Paiement à 30 jours.
```

The last two columns of the `Articles` table are always interpreted as **quantity × unit price**. Any number of columns can precede them. HT, TVA, and TTC totals are computed automatically.

### Quote

Same structure as invoice. Use `quote_number` and `validity_date` instead of `invoice_number` and `due_date`. Set `vat: 0` or disable VAT display via config to produce a HT-only total.

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

Use **ordered lists** so the numbers are visible in your source:

```markdown
## Bibliographie

1. CommonMark Spec — [spec.commonmark.org](https://spec.commonmark.org)
2. CSS Paged Media Module Level 3 — W3C Working Draft

## Sitographie

3. goldmark : [github.com/yuin/goldmark](https://github.com/yuin/goldmark)
4. paged.js : [pagedjs.org](https://pagedjs.org)
```

For sitography, any `http(s)` link in a list item is automatically moved to a new line below the entry title. Optional h3 sub-sections are supported — numbering is continuous across all sub-sections.

## Bundle authoring

Bundle templates are Go `html/template` files. The template context:

```go
type Document struct {
    Title    string
    Author   string
    Date     string
    Body     template.HTML         // rendered HTML body (do not re-escape)
    PreTOC   template.HTML         // pre-TOC sections extracted from body
    Meta     map[string]any        // all front matter fields
    TOC      []TocEntry
    Figures  []FigureEntry
    Sections map[string]Section    // h2 sections keyed by heading text
}

type Section struct {
    HTML  template.HTML
    Table [][]string               // parsed table: row 0 = headers
}

type TocEntry struct {
    Level    int
    Text     string
    ID       string
    NoNumber bool
}

type FigureEntry struct {
    ID      string                 // e.g. "fig-1"
    Caption string
}
```

The template also receives:
- `.Config` (`*config.Config`) — full config, including `.Config.CustomString "my.path" "default"` and `.Config.CustomBool "my.flag" false` for dot-path access into `custom:`
- `.ConfigCSS` (`template.HTML`) — CSS block derived from the config (font, colors, margins, …)

### Template functions

| Function | Signature | Description |
|---|---|---|
| `currency` | `float64 → string` | Formats as `1 234,56 €` |
| `toFloat` | `string → float64` | Parses numbers with spaces, commas, or `€` suffix |
| `add`, `sub`, `mul` | `float64, float64 → float64` | Arithmetic |
| `pct` | `base, rate float64 → float64` | `base × rate / 100` |
| `sumProductLast` | `[][]string → float64` | Σ(col[n-2] × col[n-1]) for each row |
| `sumProduct` | `[][]string, colA, colB int → float64` | Σ(colA[i] × colB[i]) |
| `sumCol` | `[][]string, col int → float64` | Sum of one column |
| `rowSlice` | `[][]string, from int → [][]string` | `rows[from:]` |
| `cell` | `[]string, i int → string` | `row[i]` with bounds check |
| `lastCell` | `[]string → string` | Last cell of a row |
| `prevCell` | `[]string → string` | Second-to-last cell |
| `initCells` | `[]string → []string` | All cells except the last two |

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

## Requirements

- Go 1.26.2+
- Chromium (auto-downloaded to `~/.cache/rod/` on first run)
