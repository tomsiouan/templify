---
title: Invoice
weight: 2
---

The `invoice` bundle produces a single-page invoice with automatic HT / TVA / TTC calculation.

```bash
templify -input invoice.md -bundle invoice -config invoice.yml -output invoice.pdf
```

## Front matter

```markdown
---
invoice_number: "2026-001"
date: 2026-05-10
due_date: 2026-06-09
vat: 20
---
```

## Document structure

Sections are defined as h2 headings in the Markdown file.

| Section | Description |
|---|---|
| `## Vendeur` | Vendor information (address, SIRET, …) |
| `## Client` | Client address |
| `## Articles` | Line items table |
| `## Conditions` | Payment terms, footer text |

### Items table

The last two columns are always interpreted as **quantity × unit price**. Any number of descriptive columns can precede them.

```markdown
## Articles

| Description | Unité | Qté | Prix HT |
|---|---|---|---|
| Développement API | jour | 5 | 800,00 |
| Formation | forfait | 1 | 400,00 |
```

HT, TVA, and TTC totals are computed automatically.

## Config options

```yaml
# invoice.yml
custom:
  invoice:
    logo: "./logo.svg"
    show:
      logo: true
      due_date: true
    labels:
      client: "Facturer à"
      total_ht: "Total HT"
      vat: "TVA"
      total_ttc: "Total TTC"
```
