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

	"github.com/tomsiouan/templify/pkg/document"
)

func ParseFile(path string) (*document.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	return Parse(data)
}

func Parse(data []byte) (*document.Document, error) {
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

	var toc []document.TocEntry
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !ok || !entering || h.Level > 6 {
			return ast.WalkContinue, nil
		}
		id := ""
		if v, exists := h.Attribute([]byte("id")); exists {
			id = string(v.([]byte))
		}
		toc = append(toc, document.TocEntry{Level: h.Level, Text: headingText(h, data), ID: id})
		return ast.WalkContinue, nil
	})

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, data, root); err != nil {
		return nil, fmt.Errorf("render markdown: %w", err)
	}

	metaData := meta.Get(ctx)

	doc := &document.Document{
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

	ExtractSections(doc)

	return doc, nil
}

var reImgTitle = regexp.MustCompile(`(<img\b[^>]*\btitle="([^"]*)"[^>]*>)`)

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
