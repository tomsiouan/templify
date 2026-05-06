# templify

Convert Markdown files to styled PDFs using HTML templates.

## Usage

```bash
templify -input document.md -output document.pdf
templify -input document.md -template invoice -output invoice.pdf
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-input` | — | Path to the input Markdown file (required) |
| `-output` | `output.pdf` | Path for the generated PDF |
| `-template` | `default` | Template name (from `templates/`) or path to a custom template file |

## Templates

Templates are Go HTML template files located in the `templates/` directory. A built-in `default` template is always available.

Custom templates receive a `Document` context:

```go
type Document struct {
    Title   string
    Author  string
    Date    string
    Body    template.HTML   // rendered HTML from the Markdown body
    Meta    map[string]string // any extra front matter fields
}
```

Front matter is declared as YAML at the top of the Markdown file:

```markdown
---
title: My Report
author: Alice
date: 2026-05-06
---

# Content starts here
```

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

- Go 1.21+
- A Chromium-based browser on `$PATH` (used for HTML-to-PDF rendering), **or** `wkhtmltopdf`
