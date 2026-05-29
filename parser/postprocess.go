package parser

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"github.com/tomsiouan/templify/document"
)

var reNextTopHeading = regexp.MustCompile(`<h[12][ >]`)

// ExtractPreTOC removes the named h2 sections from doc.Body and places them in
// doc.PreTOC so they can be rendered before the table of contents.
func ExtractPreTOC(doc *document.Document, headings []string) {
	if len(headings) == 0 {
		return
	}
	body := string(doc.Body)
	var extracted strings.Builder
	for _, heading := range headings {
		reHead := regexp.MustCompile(`<h2[^>]*>` + regexp.QuoteMeta(heading) + `</h2>`)
		loc := reHead.FindStringIndex(body)
		if loc == nil {
			continue
		}
		next := reNextTopHeading.FindStringIndex(body[loc[1]:])
		end := len(body)
		if next != nil {
			end = loc[1] + next[0]
		}
		extracted.WriteString(body[loc[0]:end])
		body = body[:loc[0]] + body[end:]
	}
	doc.Body = template.HTML(strings.TrimSpace(body))
	if extracted.Len() > 0 {
		doc.PreTOC = template.HTML(extracted.String())
	}
}

var (
	reFootnoteSection = regexp.MustCompile(`(?s)<div class="footnotes"[^>]*>.*?</div>`)
	reFootnoteLi      = regexp.MustCompile(`(?s)<li id="([^"]*fn:[^"]+)">\s*<p>(.*?)</p>`)
	reFootnoteBackref = regexp.MustCompile(`\s*<a[^>]+class="footnote-backref"[^>]*>.*?</a>`)
	reFootnoteRef     = regexp.MustCompile(`<sup id="[^"]*fnref:[^"]+"[^>]*><a href="#([^"]*fn:[^"]+)"[^>]*>(\d+)</a></sup>`)
)

// InlineFootnotes converts end-of-document footnotes into inline spans placed
// immediately after their reference markers, then removes the footnote section.
func InlineFootnotes(doc *document.Document) {
	body := string(doc.Body)
	section := reFootnoteSection.FindString(body)
	if section == "" {
		return
	}
	fnContent := map[string]string{}
	for _, m := range reFootnoteLi.FindAllStringSubmatch(section, -1) {
		content := reFootnoteBackref.ReplaceAllString(m[2], "")
		fnContent[m[1]] = strings.TrimSpace(content)
	}
	if len(fnContent) == 0 {
		return
	}
	body = reFootnoteRef.ReplaceAllStringFunc(body, func(s string) string {
		m := reFootnoteRef.FindStringSubmatch(s)
		if m == nil {
			return s
		}
		content, ok := fnContent[m[1]]
		if !ok {
			return s
		}
		return s + `<span class="footnote-note"><sup>` + m[2] + `</sup> ` + content + `</span>`
	})
	body = reFootnoteSection.ReplaceAllString(body, "")
	doc.Body = template.HTML(body)
}

var reHeading = regexp.MustCompile(`(?i)<(h[2-6])(\b[^>]*)>([^<]*)</h[2-6]>`)

// MarkNoNumber adds class="no-number" to headings whose text appears in exclude
// and sets NoNumber on the corresponding TOC entries, suppressing CSS counters.
func MarkNoNumber(doc *document.Document, exclude []string) {
	if len(exclude) == 0 {
		return
	}
	set := make(map[string]bool, len(exclude))
	for _, s := range exclude {
		set[strings.TrimSpace(s)] = true
	}
	body := string(doc.Body)
	body = reHeading.ReplaceAllStringFunc(body, func(s string) string {
		m := reHeading.FindStringSubmatch(s)
		if m == nil || !set[strings.TrimSpace(m[3])] {
			return s
		}
		attrs := m[2]
		if strings.Contains(attrs, `class="`) {
			attrs = strings.Replace(attrs, `class="`, `class="no-number `, 1)
		} else {
			attrs += ` class="no-number"`
		}
		return "<" + m[1] + attrs + ">" + m[3] + "</" + m[1] + ">"
	})
	doc.Body = template.HTML(body)
	for i := range doc.TOC {
		if set[strings.TrimSpace(doc.TOC[i].Text)] {
			doc.TOC[i].NoNumber = true
		}
	}
}

var (
	reFigureOpen = regexp.MustCompile(`<figure>`)
	reFigureAll  = regexp.MustCompile(`(?s)<figure>.*?</figure>`)
	reFigcap     = regexp.MustCompile(`<figcaption>([^<]*)</figcaption>`)
)

// ProcessFigures assigns sequential IDs (fig-1, fig-2, …) to <figure> elements
// that contain a <figcaption>, and populates doc.Figures with ID/caption pairs.
func ProcessFigures(doc *document.Document) {
	body := string(doc.Body)
	var figures []document.FigureEntry
	for _, fig := range reFigureAll.FindAllString(body, -1) {
		cap := reFigcap.FindStringSubmatch(fig)
		if cap == nil {
			continue
		}
		n := len(figures) + 1
		figures = append(figures, document.FigureEntry{ID: fmt.Sprintf("fig-%d", n), Caption: cap[1]})
	}
	if len(figures) == 0 {
		return
	}
	i := 0
	body = reFigureOpen.ReplaceAllStringFunc(body, func(_ string) string {
		if i >= len(figures) {
			return `<figure>`
		}
		id := figures[i].ID
		i++
		return `<figure id="` + id + `">`
	})
	doc.Body = template.HTML(body)
	doc.Figures = figures
}

// GenerateFigureTable inserts a <nav class="doc-tof"> block inside the named h2
// section, linking each entry to its figure by ID. No-op if heading is empty or
// doc.Figures is empty.
func GenerateFigureTable(doc *document.Document, heading string) {
	if heading == "" || len(doc.Figures) == 0 {
		return
	}
	reHead := regexp.MustCompile(`<h2[^>]*>` + regexp.QuoteMeta(heading) + `</h2>`)
	body := string(doc.Body)
	loc := reHead.FindStringIndex(body)
	if loc == nil {
		return
	}
	next := reNextTopHeading.FindStringIndex(body[loc[1]:])
	end := len(body)
	if next != nil {
		end = loc[1] + next[0]
	}
	var sb strings.Builder
	sb.WriteString(body[loc[0]:loc[1]])
	sb.WriteString(`<nav class="doc-tof">`)
	for _, fig := range doc.Figures {
		sb.WriteString(`<a href="#` + fig.ID + `" class="tof-entry">`)
		sb.WriteString(`<span class="tof-entry-title">` + fig.Caption + `</span>`)
		sb.WriteString(`<span class="tof-entry-filler"></span>`)
		sb.WriteString(`</a>`)
	}
	sb.WriteString(`</nav>`)
	doc.Body = template.HTML(body[:loc[0]] + sb.String() + body[end:])
}

var (
	reListItem     = regexp.MustCompile(`(?s)<li>(.*?)</li>`)
	reRefListBlock = regexp.MustCompile(`(?s)<(ul|ol)(?: start="(\d+)")?>.*?</(?:ul|ol)>`)
	reRefListOpen  = regexp.MustCompile(`^<(?:ul|ol)[^>]*>`)
	reRefListClose = regexp.MustCompile(`</(?:ul|ol)>$`)
	reSitoLink     = regexp.MustCompile(`<a\s+href="(https?://[^"]*)"[^>]*>[^<]*</a>`)
)

// NumberReferences numbers the list items in the bibliography and sitography
// sections of doc.Body. The sitography counter continues from where the
// bibliography left off.
func NumberReferences(doc *document.Document, biblioHeading, sitoHeading string) {
	body := string(doc.Body)
	counter := 0
	if biblioHeading != "" {
		body, counter = numberListSection(body, biblioHeading, counter, false)
	}
	if sitoHeading != "" {
		body, _ = numberListSection(body, sitoHeading, counter, true)
	}
	doc.Body = template.HTML(body)
}

// numberListSection numbers list items inside the named h2 section starting from
// startN. <ul> items are auto-incremented; <ol start="N"> items use N as the
// base. When isSito is true, bare links are reformatted as ref-url anchors.
// Returns the updated body and the next available counter value.
func numberListSection(body, heading string, startN int, isSito bool) (string, int) {
	reHead := regexp.MustCompile(`<h2[^>]*>` + regexp.QuoteMeta(heading) + `</h2>`)
	loc := reHead.FindStringIndex(body)
	if loc == nil {
		return body, startN
	}
	next := reNextTopHeading.FindStringIndex(body[loc[1]:])
	end := len(body)
	if next != nil {
		end = loc[1] + next[0]
	}
	section := body[loc[0]:end]

	autoN := startN
	maxUsed := startN - 1

	section = reRefListBlock.ReplaceAllStringFunc(section, func(block string) string {
		m := reRefListBlock.FindStringSubmatch(block)
		isOL := m[1] == "ol"
		olBase := 1
		if isOL && m[2] != "" {
			if n, err := strconv.Atoi(m[2]); err == nil {
				olBase = n
			}
		}
		posInList := 0
		processed := reListItem.ReplaceAllStringFunc(block, func(li string) string {
			lim := reListItem.FindStringSubmatch(li)
			if lim == nil {
				return li
			}
			var n int
			if isOL {
				n = olBase + posInList
			} else {
				n = autoN
				autoN++
			}
			posInList++
			if n > maxUsed {
				maxUsed = n
			}
			content := strings.TrimSpace(lim[1])
			if isSito {
				content = reSitoLink.ReplaceAllString(content, `<br><a class="ref-url" href="$1">$1</a>`)
			}
			return `<li class="ref-item"><span class="ref-n">[` + strconv.Itoa(n) + `]</span><span class="ref-body">` + content + `</span></li>`
		})
		processed = reRefListOpen.ReplaceAllString(processed, `<ul class="ref-list">`)
		processed = reRefListClose.ReplaceAllString(processed, `</ul>`)
		return processed
	})

	return body[:loc[0]] + section + body[end:], maxUsed + 1
}
