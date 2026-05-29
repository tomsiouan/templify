package parser

import (
	"html/template"
	"testing"

	"github.com/tomsiouan/templify/pkg/document"
)

func TestParseCells(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"two tds", "<td>foo</td><td>bar</td>", []string{"foo", "bar"}},
		{"th headers", "<th>Name</th><th>Value</th>", []string{"Name", "Value"}},
		{"strip inner tags", "<td><strong>bold</strong></td>", []string{"bold"}},
		{"trim whitespace", "<td>  hello  </td>", []string{"hello"}},
		{"mixed th and td", "<th>H</th><td>D</td>", []string{"H", "D"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseCells(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("len %d, want %d: %v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("cell[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseTable(t *testing.T) {
	t.Run("no table returns nil", func(t *testing.T) {
		if got := parseTable("<p>no table here</p>"); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("table with head and body", func(t *testing.T) {
		html := `<table>
<thead><tr><th>A</th><th>B</th></tr></thead>
<tbody>
<tr><td>1</td><td>2</td></tr>
<tr><td>3</td><td>4</td></tr>
</tbody>
</table>`
		got := parseTable(html)
		if len(got) != 3 {
			t.Fatalf("want 3 rows, got %d", len(got))
		}
		if got[0][0] != "A" || got[0][1] != "B" {
			t.Errorf("header row = %v", got[0])
		}
		if got[1][0] != "1" || got[1][1] != "2" {
			t.Errorf("data row 1 = %v", got[1])
		}
	})

	t.Run("table without thead", func(t *testing.T) {
		html := `<table><tbody><tr><td>x</td><td>y</td></tr></tbody></table>`
		got := parseTable(html)
		if len(got) != 1 {
			t.Fatalf("want 1 row, got %d", len(got))
		}
		if got[0][0] != "x" || got[0][1] != "y" {
			t.Errorf("row = %v", got[0])
		}
	})
}

func TestExtractSections(t *testing.T) {
	t.Run("no h2 produces no sections", func(t *testing.T) {
		doc := &document.Document{Body: "<p>text</p>"}
		ExtractSections(doc)
		if len(doc.Sections) != 0 {
			t.Errorf("expected no sections, got %d", len(doc.Sections))
		}
	})

	t.Run("single h2 section", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(`<h2>Items</h2><p>content</p>`)}
		ExtractSections(doc)
		if len(doc.Sections) != 1 {
			t.Fatalf("expected 1 section, got %d", len(doc.Sections))
		}
		if _, ok := doc.Sections["Items"]; !ok {
			t.Error("section 'Items' not found")
		}
	})

	t.Run("multiple h2 sections", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(`<h2>First</h2><p>a</p><h2>Second</h2><p>b</p>`)}
		ExtractSections(doc)
		if len(doc.Sections) != 2 {
			t.Fatalf("expected 2 sections, got %d", len(doc.Sections))
		}
		if _, ok := doc.Sections["First"]; !ok {
			t.Error("section 'First' not found")
		}
		if _, ok := doc.Sections["Second"]; !ok {
			t.Error("section 'Second' not found")
		}
	})

	t.Run("section with table is parsed", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(
			`<h2>Data</h2><table><tbody><tr><td>1</td><td>2</td></tr></tbody></table>`,
		)}
		ExtractSections(doc)
		s, ok := doc.Sections["Data"]
		if !ok {
			t.Fatal("section 'Data' not found")
		}
		if len(s.Table) != 1 || s.Table[0][0] != "1" {
			t.Errorf("unexpected table: %v", s.Table)
		}
	})
}
