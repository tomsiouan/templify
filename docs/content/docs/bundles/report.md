---
title: Report
weight: 1
---

The `report` bundle produces a multi-page document with a cover page, table of contents, and configurable header/footer. It is the default bundle.

```bash
templify -input document.md -bundle report -output document.pdf
```

## Front matter

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
```

`{page}` and `{pages}` are replaced with the current page number and total page count.

## Markdown features

- **GFM** — tables, strikethrough, task lists
- **Footnotes** — rendered inline at the bottom of the page where the reference appears
- **Definition lists**
- **Typographer** — smart quotes, dashes
- **Image captions** — set the `title` attribute on an image:
  ```markdown
  ![alt](image.png "This becomes a figcaption")
  ```

## Back matter

### Table of figures

Set `references.figures` in your config to the h2 heading text. The section content is replaced with an auto-generated table.

```yaml
references:
  figures: "Table des illustrations"
```

### Bibliography & sitography

```yaml
references:
  bibliography: "Bibliographie"
  sitography: "Sitographie"
```

List items are formatted as `[n]` numbered references. Use ordered lists in your source so numbers are predictable.
