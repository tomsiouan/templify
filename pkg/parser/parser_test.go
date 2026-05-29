package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestParse(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		doc, err := Parse([]byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if doc == nil {
			t.Fatal("expected non-nil document")
		}
	})

	t.Run("extracts front matter fields", func(t *testing.T) {
		src := []byte("---\ntitle: My Doc\nauthor: Alice\ndate: 2024-01-01\n---\n# Hello\n")
		doc, err := Parse(src)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if doc.Title != "My Doc" {
			t.Errorf("Title = %q, want My Doc", doc.Title)
		}
		if doc.Author != "Alice" {
			t.Errorf("Author = %q, want Alice", doc.Author)
		}
		if doc.Date != "2024-01-01" {
			t.Errorf("Date = %q, want 2024-01-01", doc.Date)
		}
	})

	t.Run("extracts headings to TOC", func(t *testing.T) {
		src := []byte("## First\n\ntext\n\n## Second\n\ntext\n\n### Sub\n")
		doc, err := Parse(src)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if len(doc.TOC) != 3 {
			t.Fatalf("TOC len = %d, want 3", len(doc.TOC))
		}
		if doc.TOC[0].Text != "First" || doc.TOC[0].Level != 2 {
			t.Errorf("TOC[0] = %+v", doc.TOC[0])
		}
		if doc.TOC[2].Level != 3 {
			t.Errorf("TOC[2].Level = %d, want 3", doc.TOC[2].Level)
		}
	})

	t.Run("renders markdown to html", func(t *testing.T) {
		src := []byte("**bold** text\n")
		doc, err := Parse(src)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if !strings.Contains(string(doc.Body), "<strong>bold</strong>") {
			t.Errorf("expected <strong> tag in body: %s", doc.Body)
		}
	})
}

func TestParseFile(t *testing.T) {
	t.Run("reads file and parses content", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "doc.md")
		if err := os.WriteFile(f, []byte("---\ntitle: File Doc\n---\n## Section\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		doc, err := ParseFile(f)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if doc.Title != "File Doc" {
			t.Errorf("Title = %q, want File Doc", doc.Title)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := ParseFile("/nonexistent/path/doc.md")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})
}

func TestWrapImageCaptions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no images unchanged", "<p>text</p>", "<p>text</p>"},
		{"img without title unchanged", `<img src="a.png" alt="desc">`, `<img src="a.png" alt="desc">`},
		{"img with empty title unchanged", `<img src="a.png" title="">`, `<img src="a.png" title="">`},
		{"img with title wrapped in figure", `<img src="a.png" title="My Figure">`, `<figure><img src="a.png" title="My Figure"><figcaption>My Figure</figcaption></figure>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := wrapImageCaptions(tc.input); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHeadingText(t *testing.T) {
	parseFirstHeading := func(src []byte) ast.Node {
		md := goldmark.New()
		root := md.Parser().Parse(text.NewReader(src))
		var h ast.Node
		_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if _, ok := n.(*ast.Heading); ok && entering {
				h = n
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		})
		return h
	}

	tests := []struct {
		name string
		src  string
		want string
	}{
		{"plain text", "## Hello\n", "Hello"},
		{"html entity unescaped", "## Foo &amp; Bar\n", "Foo & Bar"},
		{"inline code text extracted", "## Use `foo` here\n", "Use foo here"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.src)
			h := parseFirstHeading(src)
			if h == nil {
				t.Fatal("no heading found in source")
			}
			if got := headingText(h, src); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
