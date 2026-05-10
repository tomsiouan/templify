---
title: Quote
weight: 3
---

The `quote` bundle produces a single-page quote (devis). It follows the same structure as the invoice bundle but uses quote-specific fields and an optional VAT display.

```bash
templify -input quote.md -bundle quote -config quote.yml -output quote.pdf
```

## Front matter

```markdown
---
quote_number: "2026-042"
date: 2026-05-10
validity_date: 2026-06-10
vat: 20
---
```

## Document structure

| Section | Description |
|---|---|
| `## Vendeur` | Vendor information |
| `## Client` | Client address |
| `## Articles` | Line items table (same convention as invoice) |
| `## Conditions` | Validity terms, footer text |

## Config options

```yaml
# quote.yml
custom:
  quote:
    logo: "./logo.svg"
    show:
      logo: false
      validity_date: true
      vat: true          # set to false for HT-only total
    labels:
      client: "Adresser à"
      validity: "Valable jusqu'au"
      total_ht: "Total HT"
      vat: "TVA"
      total_ttc: "Total TTC"
```

When `vat: false`, the totals block shows a single HT line — useful for freelancers on franchise de base or when VAT is not applicable.
