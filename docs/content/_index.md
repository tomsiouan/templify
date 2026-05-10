---
title: templify
layout: hextra-home
---

{{< hextra/hero-badge >}}
  <span>Open Source</span>
{{< /hextra/hero-badge >}}

<div class="hx-mt-6 hx-mb-6">
{{< hextra/hero-headline >}}
  Markdown to PDF,&nbsp;<br class="sm:hx-block hx-hidden" />done right.
{{< /hextra/hero-headline >}}
</div>

<div class="hx-mb-12">
{{< hextra/hero-subtitle >}}
  Convert Markdown files to styled PDFs using a&nbsp;<br class="sm:hx-block hx-hidden" />bundle-driven template system — built in Go.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx-mb-6">
{{< hextra/hero-button text="Get Started" link="docs/getting-started" >}}
</div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Bundle-driven"
    subtitle="Pick a built-in bundle (report, invoice, quote) or bring your own directory with a custom template."
  >}}
  {{< hextra/feature-card
    title="Semantic Markdown"
    subtitle="H2 sections become structured data units — tables in Markdown become typed rows in the template."
  >}}
  {{< hextra/feature-card
    title="Config-layered"
    subtitle="Default → bundle → config file → front matter. Override only what you need at each level."
  >}}
  {{< hextra/feature-card
    title="No install required"
    subtitle="Chromium is auto-downloaded on first run. One binary, no dependencies."
  >}}
  {{< hextra/feature-card
    title="Go templates"
    subtitle="Full html/template power with built-in math helpers for totals, currencies, and table operations."
  >}}
  {{< hextra/feature-card
    title="CSS Paged Media"
    subtitle="Page breaks, margin boxes, and print-ready output via paged.js — no manual layout tuning."
  >}}
{{< /hextra/feature-grid >}}
