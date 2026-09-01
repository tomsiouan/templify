---
title: Configuration
weight: 3
---

Configuration is layered — each level overrides the previous:

1. Built-in defaults
2. Bundle `default.yml`
3. User `-config` file
4. Front matter (header/footer slots only)

## Full reference

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
  # faces:                        # self-host instead of linking url, see "Self-hosting fonts"
  #   - file: ./fonts/inter-400.woff2
  #     weight: 400
  size: 11pt
  line_height: 1.6

justify: false
paragraph_indent: 13mm
heading_indent: 5mm

heading_numbers:
  enabled: true
  exclude:
    - Introduction
    - Conclusion

toc:
  enabled: true
  max_depth: 3
  exclude:
    - Remerciements
  pre_toc:
    - Remerciements

references:
  bibliography: "Bibliographie"
  sitography: "Sitographie"
  figures: "Table des illustrations"

header:
  enabled: true
  background: "linear-gradient(to right, #1e293b, #475569)"
  left: ""
  center: ""
  right: ""

footer:
  enabled: true
  background: ""
  left: ""
  center: ""
  right: "{page} / {pages}"

blank_page: false

colors:
  primary: "#1e293b"
  primary_light: "#475569"
  background: "#f8fafc"
  text: "#0f172a"
  text_muted: "#64748b"

code:
  theme: monokai
  background: ""
  foreground: ""
  font_family: JetBrains Mono
  font_url: https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap
  font_size: 8.5pt
  line_height: 1.45
  line_numbers: false
  wrap: true
  keep_together_lines: 25

cover:
  enabled: true
  template: ""

custom: {}
```

## Self-hosting fonts

By default, `font.url` and `code.font_url` link a webfont stylesheet — Google Fonts, for the built-in bundles — and Chromium fetches it while rendering the page. This works, but the PDF's text layer can come out wrong in a way that only shows up when you copy text out of it: letters transposed, words split by stray spaces, lines run together. Chromium's headless PDF export mis-handles certain webfonts, most commonly *variable fonts* — Google Fonts hands one out whenever a stylesheet URL requests more than one weight (`wght@400;600`), even though each individual `@font-face` rule in that stylesheet only claims one of those weights.

`faces` sidesteps the whole class of bug: it self-hosts a **static** font file per weight/style, inlined into the document as a data URI instead of linked.

```yaml
font:
  family: Inter
  faces:
    - file: ./fonts/inter-400.woff2
      weight: 400
    - file: ./fonts/inter-600.woff2
      weight: 600
      unicode_range: "U+0000-00FF" # optional, passed through as-is
```

`file` resolves relative to the config file. `weight` defaults to `400`, `style` to `normal`. Setting `faces` drops the matching `url`: loading the same family both linked and inlined would put the bug back. The same `faces` key exists under `code:`, for the code font.

### Fetching faces automatically

```bash
templify -fetch-fonts -config myconfig.yml
```

reads `font.url` and `code.font_url`, downloads every file their stylesheet references, and prints a `faces:` block for each — paste it into the config in place of the `url:` line. Files land in `./fonts` next to the config by default; `-fonts-dir` overrides that.

When the source is Google Fonts and a downloaded file turns out to be a variable font, `-fetch-fonts` re-requests that weight on its own so the file it saves is a static instance — this is the actual fix, applied automatically for the common case. A variable file from any other host is kept, with a warning: there is no general way to turn an arbitrary third-party variable font into a static one.

The config itself is never rewritten — yaml.v2 cannot round-trip comments, and these files usually have some — so the block is printed for you to paste in by hand.

This is a one-time step (or "whenever the fonts change"), not something that runs on every render. Commit the fetched files alongside the config, and `templify` never needs network access to produce a PDF.

### The other half of the fix: whole-pixel font sizes

Self-hosting a static font removes the worst of the corruption — variable fonts break almost every word — but a narrower version of the same symptom can still show up around specific glyphs, most visibly the underscore in monospace code (`disable_mlock` copying out as `disable` / `_` / `mlock` on separate lines), even with a confirmed-static, correctly embedded font.

The cause: Chromium prints any `@font-face`-loaded text as one positioned run per glyph rather than one run per word, and if the font-size computes to a fractional device pixel, consecutive glyph advances end up inconsistent by a sub-pixel amount. Apple's PDFKit (Preview, Quick Look) mis-orders text around that inconsistency. A completely independent engine (MuPDF) reads the exact same PDF perfectly — the file is correct; the reader's reconstruction of copyable text is what fails.

templify wraps every generated `font-size` — body text, headings, the code panel, inline code — in CSS's `round(value, 1px)`, so it always lands on a whole pixel regardless of what pt/em value is configured. This is automatic and requires no configuration; `round()` is a no-op when the value is already whole, so it only ever helps. It needs Chromium 114+, which every browser templify launches (system-installed or its own downloaded copy) comfortably clears.

Self-hosting and whole-pixel rounding fix two different triggers of the same class of bug and neither replaces the other: rounding alone does not fix a variable font's internally-interpolated glyph metrics, and a static font alone does not fix a fractional-pixel font-size. Both are needed for fully clean copy-paste.

### A third trigger: the font's own unitsPerEm

Even with a static font and a whole-pixel font-size, one specific font can still corrupt copy-paste around one specific glyph, on every occurrence, regardless of size. The cause is inside the font file itself: a glyph's advance is stored as a fraction of the font's `unitsPerEm` (e.g. Manrope's underscore is 1320/2000 = 0.66 em). Scaling that fraction to *any* pixel size lands on a whole pixel only when `unitsPerEm` is a power of two (1024, 2048, 4096 — the convention most TrueType fonts follow) — binary floating point represents that division cleanly. A `unitsPerEm` of 1000 or 2000 (common among fonts derived from PostScript/CFF sources — Manrope and JetBrains Mono both use it) makes the division inexact, and PDFKit's misordering triggers again, immune to `round(font-size, 1px)` because the corruption doesn't come from the size at all.

Check a font's `unitsPerEm` with [fontTools](https://github.com/fonttools/fonttools) (`pip install fonttools brotli`):

```python
from fontTools.ttLib import TTFont
print(TTFont("your-font.woff2")["head"].unitsPerEm)
```

If it isn't a power of two, rescale it — this uniformly resizes every outline and metric, so the font looks identical, just represented at a different internal resolution:

```python
from fontTools.ttLib import TTFont
from fontTools.ttLib.scaleUpem import scale_upem

f = TTFont("your-font.woff2")
scale_upem(f, 2048)
f.flavor = "woff2"
f.save("your-font-2048.woff2")
```

Do this once per font file, point `faces:` at the rescaled copies, and the glyph-level corruption is gone regardless of font-size. templify does not do this automatically — WOFF2 table rewriting needs a Brotli codec templify does not currently depend on — so for now it is a manual, one-time step per font.

### A fourth trigger: padding on inline elements

`padding` on an inline element sitting inside justified, reflowing prose (templify's inline `code` styling, before this was found) can make Chromium compute a slightly different line-box height for the one line carrying it, reintroducing the same fractional-position corruption for that line — occasionally swapping it with its wrapped neighbor on copy. templify's inline code now gets its highlight from `box-shadow`'s spread instead of `padding`: `box-shadow` paints outside the element's box without taking part in layout, so it looks identical but never perturbs the line box. If you add padding to your own inline (not block) elements inside body text in a custom bundle, prefer `box-shadow: 0 0 0 <spread> <color>` the same way.

## Code blocks

Fenced code blocks are rendered as a filled, rounded panel with syntax
highlighting, so they never read as a blockquote. The language comes from the
fence itself:

````markdown
```go
func main() {}
```
````

`code.theme` names a [Chroma](https://github.com/alecthomas/chroma) style. Run
`templify -code-themes` to list the 70+ available names, or set `none` to keep
the panel without coloring. The default is `monokai`.

`background` and `foreground` override just the panel colors while keeping the
theme's token colors. That is the knob for a theme whose colors you like on a
panel you don't: `monokai` sits on a near-black `#272822`, and printers handle a
lighter panel better, so `background: "#32332b"` keeps the palette on less ink.
It is also the fix for a theme whose comment color is too dim to print — one of
`monokai`'s weak spots, at `#75715e` on its own background.

### Turning the panel off

`background: none` (or `transparent`) drops the filled panel and its padding,
leaving code directly on the page. Pair it with a light theme, because a dark
theme's token colors were chosen to sit on that dark background and wash out on
white — templify logs a warning if you combine the two:

```yaml
code:
  theme: github
  background: none
```

With the panel off and no explicit `foreground`, code takes the body text color
rather than the theme's, for the same reason.

### Line numbers

`line_numbers` is off by default. Numbers are drawn by the highlighter, so they
only appear on fences that name a language it recognises — an untagged fence
renders plain, with no gutter. Tag it `text` to get numbers without coloring:

````markdown
```text
no language, still numbered
```
````

`wrap: true` folds long lines inside the panel instead of letting them run off
the page, breaking only tokens that cannot fit on a line of their own. Wrapped
lines hang under the gutter when line numbers are on.

`keep_together_lines` is the cutoff for keeping a block whole: blocks up to that
many lines never split across a page boundary, while longer ones are allowed to
break. Blocks taller than the text area cannot be kept whole at all, so forcing
them risks losing content — set `0` to let every block break.

## Colors as CSS variables

| Variable | Config key |
|---|---|
| `--blue` | `colors.primary` |
| `--blue-light` | `colors.primary_light` |
| `--bg-soft` | `colors.background` |
| `--text` | `colors.text` |
| `--text-muted` | `colors.text_muted` |
| `--code-bg` | `code.background`, else the theme's background |
| `--code-fg` | `code.foreground`, else the theme's text color |
| `--code-font` | `code.font_family` plus monospace fallbacks |
| `--code-font-size` | `code.font_size` |
| `--code-line-height` | `code.line_height` |

## Bundle-specific options

Options under `custom:` are accessible in templates via `.Config.CustomString` and `.Config.CustomBool`:

```yaml
custom:
  invoice:
    show:
      logo: true
    logo: ./logo.svg
```

```html
{{- $showLogo := .Config.CustomBool "invoice.show.logo" false -}}
{{- $logoPath  := .Config.CustomString "invoice.logo" "" -}}
```
