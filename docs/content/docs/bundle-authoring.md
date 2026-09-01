---
title: Bundle Authoring
weight: 4
---

Bundle templates are Go [`html/template`](https://pkg.go.dev/html/template) files.

## Template context

```go
type Document struct {
    Title    string
    Author   string
    Date     string
    Body     template.HTML         // rendered HTML body — do not re-escape
    PreTOC   template.HTML         // pre-TOC sections extracted from body
    Meta     map[string]any        // all front matter fields
    TOC      []TocEntry
    Figures  []FigureEntry
    Sections map[string]Section    // h2 sections keyed by heading text
}

type Section struct {
    HTML  template.HTML
    Table [][]string               // row 0 = headers, row 1+ = data
}
```

`.Config` (`*config.Config`) and `.ConfigCSS` (`template.HTML`) are also available on every render.

## Accessing sections

```html
{{with index .Sections "Client"}}
  <div>{{.HTML}}</div>
{{end}}

{{- $articles := index .Sections "Articles" -}}
{{- $headers  := index $articles.Table 0 -}}
{{- $rows     := rowSlice $articles.Table 1 -}}
```

## Template functions

### Table helpers

| Function | Description |
|---|---|
| `rowSlice rows from` | `rows[from:]` |
| `cell row i` | `row[i]` with bounds check |
| `lastCell row` | Last cell of a row |
| `prevCell row` | Second-to-last cell |
| `initCells row` | All cells except the last two |

### Math

| Function | Signature | Description |
|---|---|---|
| `toFloat` | `string → float64` | Parses numbers with spaces, commas, or `€` suffix |
| `add` | `float64, float64 → float64` | Addition |
| `sub` | `float64, float64 → float64` | Subtraction |
| `mul` | `float64, float64 → float64` | Multiplication |
| `pct` | `base, rate → float64` | `base × rate / 100` |
| `sumProductLast` | `[][]string → float64` | Σ(col[n-2] × col[n-1]) — last two columns |
| `sumProduct` | `[][]string, colA, colB → float64` | Σ(colA[i] × colB[i]) |
| `sumCol` | `[][]string, col → float64` | Sum of one column |

### Formatting

| Function | Description |
|---|---|
| `currency f` | Formats as `1 234,56 €` |

## Minimal example

```html
<!doctype html>
<html>
<head>
  {{.ConfigCSS}}
</head>
<body>
  <h1>{{.Title}}</h1>
  {{.Body}}
</body>
</html>
```

## Styling code blocks

`{{.ConfigCSS}}` styles code blocks from the [`code:` config
section](../configuration/#code-blocks), and those rules are prefixed with
`html` so they outrank a plain `pre { … }` rule wherever it sits in your
stylesheet. That is deliberate: a theme's token colors and its panel background
are two halves of one thing, and a bundle that repaints only the background
leaves light text on a light panel.

So do not style `pre`, `pre code` or `code` in your bundle. Point users at the
`code:` options instead, or override the CSS variables:

```html
<style>
  :root {
    --code-bg: #1c1f26;
    --code-font-size: 9pt;
  }
</style>
```

If you need the panel itself, match the prefix so your rule wins, and set
`code.theme` to `none` in `default.yml` so no token colors are left stranded:

```html
<style>
  html pre { background: #eef1f6; color: #1c1f26; }
</style>
```

This same reasoning extends beyond code: avoid `padding` on any inline
(not block) element you place inside body text — inline `code` is the one
built-in example. Padding on an inline element sitting inside reflowing text
can make Chromium compute a slightly different line-box height for the one
line carrying it, which some PDF readers occasionally reorder relative to its
neighbor on copy (see [Self-hosting fonts](../configuration/#self-hosting-fonts)
for the full story). `box-shadow: 0 0 0 <spread> <color>` gives the same visual
highlight — its spread paints outside the box without taking part in layout —
with none of that risk.

Bundles written before the `code:` section usually carry a copy of the old
quote-style `pre` rules. Delete them: they are dead weight now, and the block
would otherwise be styled in two places at once.
