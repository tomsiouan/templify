package parser

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type Document struct {
	Title  string
	Author string
	Date   string
	Body   template.HTML
	Meta   map[string]any
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
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := md.Convert(data, &buf, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("convert markdown: %w", err)
	}

	metaData := meta.Get(ctx)

	doc := &Document{
		Body: template.HTML(buf.String()),
		Meta: metaData,
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
