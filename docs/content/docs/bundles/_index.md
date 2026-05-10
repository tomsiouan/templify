---
title: Bundles
weight: 2
---

A bundle is a directory containing:

- `main.html` — Go `html/template` for the document body (required)
- `cover.html` — optional cover page template
- `default.yml` — optional config defaults specific to this bundle

Built-in bundles are embedded in the binary. You can also use a local directory via `-bundle ./my-bundle/`.
