package parser

import (
	"bytes"
	"fmt"
	gohtml "html"
	"html/template"
	"os"
	"regexp"
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

type Document struct {
	Title  string
	Author string
	Date   string
	Body   template.HTML
	PreTOC template.HTML // sections extracted from body to render before the TOC
	Meta   map[string]any
	TOC    []TocEntry
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
