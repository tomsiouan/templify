package parser

import (
	"strings"
	"testing"

)

func TestWrapImageCaptions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"no images unchanged",
			"<p>text</p>",
			"<p>text</p>",
		},
		{
			"img without title unchanged",
			`<img src="a.png" alt="desc">`,
			`<img src="a.png" alt="desc">`,
		},
		{
			"img with empty title unchanged",
			`<img src="a.png" title="">`,
			`<img src="a.png" title="">`,
		},
		{
			"img with title wrapped in figure",
			`<img src="a.png" title="My Figure">`,
			`<figure><img src="a.png" title="My Figure"><figcaption>My Figure</figcaption></figure>`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := wrapImageCaptions(tc.input); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

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
