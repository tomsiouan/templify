---
title: Custom Bundles
weight: 4
---

You can use a local directory as a bundle. Pass its path to `-bundle`:

```bash
templify -input document.md -bundle ./my-bundle/ -output document.pdf
```

## Directory structure

```
my-bundle/
├── main.html       # required — Go html/template
├── cover.html      # optional cover page
└── default.yml     # optional config defaults
```

## default.yml

Values in `default.yml` are layered on top of the global defaults and under the user's `-config` file. Only set values that differ from the defaults.

```yaml
cover:
  enabled: false

page:
  size: A4
  margins:
    top: 14mm
    right: 14mm
    bottom: 14mm
    left: 14mm

custom:
  mybundle:
    some_option: "value"
```

## Cover page

If `cover.html` is present and `cover.enabled: true`, it is rendered as a separate named page. Any `<style>` blocks inside the cover template are automatically extracted and injected into `<head>` to avoid blank pages.
