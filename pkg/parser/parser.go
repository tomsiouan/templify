package parser

import (
	"bytes"
	"fmt"
	gohtml "html"
	"html/template"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

type TocEntry struct {
	Level    int
	Text     string
	ID       string
	NoNumber bool
}

type FigureEntry struct {
	ID      string
	Caption string
}

type Document struct {
	Title   string
	Author  string
	Date    string
	Body    template.HTML
	PreTOC  template.HTML // sections extracted from body to render before the TOC
	Meta    map[string]any
	TOC     []TocEntry
	Figures []FigureEntry
}

func ParseFile(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	return Parse(data)
}

func Parse(data []byte) (*Document, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			extension.GFM,
			extension.Table,
			extension.Footnote,
			extension.DefinitionList,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	ctx := parser.NewContext()
	reader := text.NewReader(data)
	root := md.Parser().Parse(reader, parser.WithContext(ctx))

	var toc []TocEntry
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !ok || !entering || h.Level > 6 {
			return ast.WalkContinue, nil
		}
		id := ""
		if v, exists := h.Attribute([]byte("id")); exists {
			id = string(v.([]byte))
		}
		toc = append(toc, TocEntry{Level: h.Level, Text: headingText(h, data), ID: id})
		return ast.WalkContinue, nil
	})

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, data, root); err != nil {
		return nil, fmt.Errorf("render markdown: %w", err)
	}

	metaData := meta.Get(ctx)

	doc := &Document{
		Body: template.HTML(wrapImageCaptions(buf.String())),
		Meta: metaData,
		TOC:  toc,
	}

	if v, ok := metaData["title"]; ok {
		doc.Title, _ = v.(string)
	}
	if v, ok := metaData["author"]; ok {
		doc.Author, _ = v.(string)
	}
	if v, ok := metaData["date"]; ok {
		doc.Date, _ = v.(string)
	}

	return doc, nil
}

var reNextTopHeading = regexp.MustCompile(`<h[12][ >]`)

// ExtractPreTOC removes sections matching the given headings from doc.Body and
// stores their HTML in doc.PreTOC, in the order they appear in the document.
func ExtractPreTOC(doc *Document, headings []string) {
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
		// find the next h1/h2 after this heading to determine section end
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
	reImgTitle        = regexp.MustCompile(`(<img\b[^>]*\btitle="([^"]*)"[^>]*>)`)
	reFootnoteSection = regexp.MustCompile(`(?s)<div class="footnotes"[^>]*>.*?</div>`)
	reFootnoteLi      = regexp.MustCompile(`(?s)<li id="([^"]*fn:[^"]+)">\s*<p>(.*?)</p>`)
	reFootnoteBackref = regexp.MustCompile(`\s*<a[^>]+class="footnote-backref"[^>]*>.*?</a>`)
	reFootnoteRef     = regexp.MustCompile(`<sup id="[^"]*fnref:[^"]+"[^>]*><a href="#([^"]*fn:[^"]+)"[^>]*>(\d+)</a></sup>`)

	reFigureOpen    = regexp.MustCompile(`<figure>`)
	reFigureAll     = regexp.MustCompile(`(?s)<figure>.*?</figure>`)
	reFigcap        = regexp.MustCompile(`<figcaption>([^<]*)</figcaption>`)
	reListItem      = regexp.MustCompile(`(?s)<li>(.*?)</li>`)
	reRefListBlock  = regexp.MustCompile(`(?s)<(ul|ol)(?: start="(\d+)")?>.*?</(?:ul|ol)>`)
	reRefListOpen   = regexp.MustCompile(`^<(?:ul|ol)[^>]*>`)
	reRefListClose  = regexp.MustCompile(`</(?:ul|ol)>$`)
	reSitoLink      = regexp.MustCompile(`<a\s+href="(https?://[^"]*)"[^>]*>[^<]*</a>`)
)

// InlineFootnotes moves footnote content from the end-of-document section to
// inline spans right after each reference superscript, then removes the section.
func InlineFootnotes(doc *Document) {
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

// MarkNoNumber adds class="no-number" to h2-h6 headings whose text matches
// the exclude list, and sets NoNumber on the corresponding TOC entries.
func MarkNoNumber(doc *Document, exclude []string) {
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

// wrapImageCaptions wraps <img title="..."> with <figure><figcaption> when a title is present.
func wrapImageCaptions(html string) string {
	return reImgTitle.ReplaceAllStringFunc(html, func(img string) string {
		m := reImgTitle.FindStringSubmatch(img)
		if m[2] == "" {
			return img
		}
		return "<figure>" + m[1] + "<figcaption>" + m[2] + "</figcaption></figure>"
	})
}

// ProcessFigures assigns sequential IDs to <figure> elements and populates doc.Figures.
func ProcessFigures(doc *Document) {
	body := string(doc.Body)
	var figures []FigureEntry
	for _, fig := range reFigureAll.FindAllString(body, -1) {
		cap := reFigcap.FindStringSubmatch(fig)
		if cap == nil {
			continue
		}
		n := len(figures) + 1
		figures = append(figures, FigureEntry{ID: fmt.Sprintf("fig-%d", n), Caption: cap[1]})
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

// GenerateFigureTable replaces the content of the given h2 section with an
// auto-generated table of figures (dotted leaders; page numbers filled by JS).
func GenerateFigureTable(doc *Document, heading string) {
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

// NumberReferences auto-numbers list items in bibliography and sitography sections.
// Sitography numbering continues from where bibliography left off.
// For sitography items, http(s) links are moved to a new line.
func NumberReferences(doc *Document, biblioHeading, sitoHeading string) {
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

	// <ol>/<ol start="N"> items use the list's own numbering;
	// <ul> items fall back to a global auto-counter.
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

func headingText(n ast.Node, source []byte) string {
	var sb strings.Builder
	_ = ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := node.(type) {
		case *ast.Text:
			sb.Write(v.Segment.Value(source))
		case *ast.String:
			sb.Write(v.Value)
		}
		return ast.WalkContinue, nil
	})
	return gohtml.UnescapeString(strings.TrimSpace(sb.String()))
}
