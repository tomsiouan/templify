package parser

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

type TocEntry struct {
	Level int
	Text  string
	ID    string
}

type Document struct {
	Title  string
	Author string
	Date   string
	Body   template.HTML
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
	_ = gast.Walk(root, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		h, ok := n.(*gast.Heading)
		if !ok || !entering || h.Level > 3 {
			return gast.WalkContinue, nil
		}
		id := ""
		if v, exists := h.Attribute([]byte("id")); exists {
			id = string(v.([]byte))
		}
		toc = append(toc, TocEntry{Level: h.Level, Text: headingText(h, data), ID: id})
		return gast.WalkContinue, nil
	})

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, data, root); err != nil {
		return nil, fmt.Errorf("render markdown: %w", err)
	}

	metaData := meta.Get(ctx)

	doc := &Document{
		Body: template.HTML(buf.String()),
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

func headingText(n gast.Node, source []byte) string {
	var sb strings.Builder
	_ = gast.Walk(n, func(node gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}
		switch v := node.(type) {
		case *gast.Text:
			sb.Write(v.Segment.Value(source))
		case *gast.String:
			sb.Write(v.Value)
		}
		return gast.WalkContinue, nil
	})
	return strings.TrimSpace(sb.String())
}
