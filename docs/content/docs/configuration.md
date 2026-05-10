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

cover:
  enabled: true
  template: ""

custom: {}
```

## Colors as CSS variables

| Variable | Config key |
|---|---|
| `--blue` | `colors.primary` |
| `--blue-light` | `colors.primary_light` |
| `--bg-soft` | `colors.background` |
| `--text` | `colors.text` |
| `--text-muted` | `colors.text_muted` |

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
