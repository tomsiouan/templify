package parser

import (
	"html/template"
	"strings"
	"testing"

	"github.com/tomsiouan/templify/pkg/document"
)

func TestNumberListSection(t *testing.T) {
	t.Run("section not found leaves body unchanged", func(t *testing.T) {
		body := "<p>text</p>"
		got, n := numberListSection(body, "References", 0, false)
		if got != body {
			t.Errorf("body changed unexpectedly: %q", got)
		}
		if n != 0 {
			t.Errorf("counter = %d, want 0", n)
		}
	})

	t.Run("ul items numbered from startN", func(t *testing.T) {
		body := `<h2>Refs</h2><ul><li>First</li><li>Second</li></ul>`
		got, n := numberListSection(body, "Refs", 0, false)
		if !strings.Contains(got, `[0]`) || !strings.Contains(got, `[1]`) {
			t.Errorf("expected [0] and [1] in: %s", got)
		}
		if n != 2 {
			t.Errorf("next counter = %d, want 2", n)
		}
	})

	t.Run("ol respects start attribute", func(t *testing.T) {
		body := `<h2>Refs</h2><ol start="5"><li>Item</li></ol>`
		got, _ := numberListSection(body, "Refs", 0, false)
		if !strings.Contains(got, `[5]`) {
			t.Errorf("expected [5] in: %s", got)
		}
	})

	t.Run("counter chains across two sections", func(t *testing.T) {
		body := `<h2>Biblio</h2><ul><li>A</li><li>B</li></ul><h2>Sito</h2><ul><li>C</li></ul>`
		var n int
		body, n = numberListSection(body, "Biblio", 0, false)
		body, _ = numberListSection(body, "Sito", n, true)
		if !strings.Contains(body, `[2]`) {
			t.Errorf("sito should start at [2]: %s", body)
		}
	})

	t.Run("sito adds ref-url class to links", func(t *testing.T) {
		body := `<h2>Sito</h2><ul><li><a href="https://example.com">Example</a></li></ul>`
		got, _ := numberListSection(body, "Sito", 0, true)
		if !strings.Contains(got, `class="ref-url"`) {
			t.Errorf("expected ref-url class in: %s", got)
		}
		if !strings.Contains(got, "https://example.com") {
			t.Errorf("expected URL in: %s", got)
		}
	})
}

func TestMarkNoNumber(t *testing.T) {
	t.Run("empty exclude list leaves body unchanged", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(`<h2>Title</h2>`)}
		MarkNoNumber(doc, nil)
		if strings.Contains(string(doc.Body), "no-number") {
			t.Errorf("unexpected no-number class: %s", doc.Body)
		}
	})

	t.Run("matching heading gets no-number class and TOC flag", func(t *testing.T) {
		doc := &document.Document{
			Body: template.HTML(`<h2>Appendix</h2>`),
			TOC:  []document.TocEntry{{Level: 2, Text: "Appendix"}},
		}
		MarkNoNumber(doc, []string{"Appendix"})
		if !strings.Contains(string(doc.Body), `class="no-number"`) {
			t.Errorf("expected no-number class: %s", doc.Body)
		}
		if !doc.TOC[0].NoNumber {
			t.Error("TOC entry should be marked NoNumber")
		}
	})

	t.Run("existing class gets no-number prepended", func(t *testing.T) {
		doc := &document.Document{
			Body: template.HTML(`<h2 class="existing">Title</h2>`),
		}
		MarkNoNumber(doc, []string{"Title"})
		if !strings.Contains(string(doc.Body), `class="no-number existing"`) {
			t.Errorf("expected no-number prepended: %s", doc.Body)
		}
	})

	t.Run("unmatched heading is not modified", func(t *testing.T) {
		doc := &document.Document{
			Body: template.HTML(`<h2>Keep</h2>`),
			TOC:  []document.TocEntry{{Level: 2, Text: "Keep"}},
		}
		MarkNoNumber(doc, []string{"Other"})
		if strings.Contains(string(doc.Body), "no-number") {
			t.Errorf("unexpected no-number class: %s", doc.Body)
		}
		if doc.TOC[0].NoNumber {
			t.Error("TOC entry should not be marked")
		}
	})
}

func TestProcessFigures(t *testing.T) {
	t.Run("no figures leaves doc.Figures nil", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML("<p>no figures</p>")}
		ProcessFigures(doc)
		if len(doc.Figures) != 0 {
			t.Errorf("expected no figures, got %d", len(doc.Figures))
		}
	})

	t.Run("figure without figcaption is skipped", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(`<figure><img src="x.png"></figure>`)}
		ProcessFigures(doc)
		if len(doc.Figures) != 0 {
			t.Errorf("expected no figures, got %d", len(doc.Figures))
		}
	})

	t.Run("single figure gets id and caption", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(
			`<figure><img src="x.png"><figcaption>My Caption</figcaption></figure>`,
		)}
		ProcessFigures(doc)
		if len(doc.Figures) != 1 {
			t.Fatalf("expected 1 figure, got %d", len(doc.Figures))
		}
		if doc.Figures[0].ID != "fig-1" {
			t.Errorf("ID = %q, want fig-1", doc.Figures[0].ID)
		}
		if doc.Figures[0].Caption != "My Caption" {
			t.Errorf("Caption = %q, want My Caption", doc.Figures[0].Caption)
		}
		if !strings.Contains(string(doc.Body), `id="fig-1"`) {
			t.Errorf("body missing id: %s", doc.Body)
		}
	})

	t.Run("multiple figures get sequential IDs", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(
			`<figure><img src="a.png"><figcaption>Cap A</figcaption></figure>` +
				`<figure><img src="b.png"><figcaption>Cap B</figcaption></figure>`,
		)}
		ProcessFigures(doc)
		if len(doc.Figures) != 2 {
			t.Fatalf("expected 2 figures, got %d", len(doc.Figures))
		}
		if doc.Figures[0].ID != "fig-1" || doc.Figures[1].ID != "fig-2" {
			t.Errorf("IDs = %q, %q", doc.Figures[0].ID, doc.Figures[1].ID)
		}
	})
}

func TestInlineFootnotes(t *testing.T) {
	t.Run("no footnote section leaves body unchanged", func(t *testing.T) {
		body := "<p>text</p>"
		doc := &document.Document{Body: template.HTML(body)}
		InlineFootnotes(doc)
		if string(doc.Body) != body {
			t.Errorf("unexpected change: %s", doc.Body)
		}
	})

	t.Run("footnote inlined and section removed", func(t *testing.T) {
		body := `<p>text<sup id="fnref:1"><a href="#fn:1">1</a></sup></p>` +
			`<div class="footnotes" role="doc-endnotes">` +
			`<ol><li id="fn:1"><p>Note content <a href="#fnref:1" class="footnote-backref">↩</a></p></li></ol>` +
			`</div>`
		doc := &document.Document{Body: template.HTML(body)}
		InlineFootnotes(doc)
		if strings.Contains(string(doc.Body), `class="footnotes"`) {
			t.Error("footnote section should be removed from body")
		}
		if !strings.Contains(string(doc.Body), `class="footnote-note"`) {
			t.Errorf("expected inline footnote span: %s", doc.Body)
		}
		if !strings.Contains(string(doc.Body), "Note content") {
			t.Errorf("footnote content missing: %s", doc.Body)
		}
	})
}

func TestGenerateFigureTable(t *testing.T) {
	t.Run("empty heading skipped", func(t *testing.T) {
		doc := &document.Document{
			Body:    template.HTML(`<h2>Figures</h2>`),
			Figures: []document.FigureEntry{{ID: "fig-1", Caption: "Cap"}},
		}
		GenerateFigureTable(doc, "")
		if strings.Contains(string(doc.Body), "doc-tof") {
			t.Errorf("should not generate tof with empty heading: %s", doc.Body)
		}
	})

	t.Run("no figures skipped", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(`<h2>Figures</h2>`)}
		GenerateFigureTable(doc, "Figures")
		if strings.Contains(string(doc.Body), "doc-tof") {
			t.Errorf("should not generate tof with no figures: %s", doc.Body)
		}
	})

	t.Run("heading not found skipped", func(t *testing.T) {
		doc := &document.Document{
			Body:    template.HTML(`<h2>Other</h2>`),
			Figures: []document.FigureEntry{{ID: "fig-1", Caption: "Cap"}},
		}
		GenerateFigureTable(doc, "Figures")
		if strings.Contains(string(doc.Body), "doc-tof") {
			t.Errorf("should not generate tof when heading not found: %s", doc.Body)
		}
	})

	t.Run("inserts tof nav with figure links", func(t *testing.T) {
		doc := &document.Document{
			Body: template.HTML(`<h2>Figures</h2><p>placeholder</p>`),
			Figures: []document.FigureEntry{
				{ID: "fig-1", Caption: "Cap A"},
				{ID: "fig-2", Caption: "Cap B"},
			},
		}
		GenerateFigureTable(doc, "Figures")
		if !strings.Contains(string(doc.Body), `class="doc-tof"`) {
			t.Errorf("expected doc-tof nav: %s", doc.Body)
		}
		if !strings.Contains(string(doc.Body), `href="#fig-1"`) {
			t.Errorf("expected link to fig-1: %s", doc.Body)
		}
		if !strings.Contains(string(doc.Body), "Cap B") {
			t.Errorf("expected caption Cap B: %s", doc.Body)
		}
	})
}

func TestNumberReferences(t *testing.T) {
	t.Run("numbers bibliography only", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(
			`<h2>Refs</h2><ul><li>A</li><li>B</li></ul>`,
		)}
		NumberReferences(doc, "Refs", "")
		if !strings.Contains(string(doc.Body), `[0]`) || !strings.Contains(string(doc.Body), `[1]`) {
			t.Errorf("expected [0] and [1]: %s", doc.Body)
		}
	})

	t.Run("numbers sitography continuing from bibliography", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(
			`<h2>Biblio</h2><ul><li>A</li></ul><h2>Sito</h2><ul><li>B</li></ul>`,
		)}
		NumberReferences(doc, "Biblio", "Sito")
		if !strings.Contains(string(doc.Body), `[0]`) {
			t.Errorf("expected [0] in biblio: %s", doc.Body)
		}
		if !strings.Contains(string(doc.Body), `[1]`) {
			t.Errorf("expected [1] in sito: %s", doc.Body)
		}
	})

	t.Run("empty headings leave body unchanged", func(t *testing.T) {
		body := `<h2>Section</h2><ul><li>A</li></ul>`
		doc := &document.Document{Body: template.HTML(body)}
		NumberReferences(doc, "", "")
		if string(doc.Body) != body {
			t.Errorf("body changed unexpectedly: %s", doc.Body)
		}
	})
}

func TestExtractPreTOC(t *testing.T) {
	t.Run("empty headings list leaves body unchanged", func(t *testing.T) {
		body := "<h2>Section</h2><p>text</p>"
		doc := &document.Document{Body: template.HTML(body)}
		ExtractPreTOC(doc, nil)
		if string(doc.Body) != body {
			t.Errorf("body changed unexpectedly: %s", doc.Body)
		}
	})

	t.Run("extracted section removed from body and placed in PreTOC", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(
			`<h2>Intro</h2><p>intro text</p><h2>Main</h2><p>main text</p>`,
		)}
		ExtractPreTOC(doc, []string{"Intro"})
		if strings.Contains(string(doc.Body), "<h2>Intro</h2>") {
			t.Error("Intro section should be removed from body")
		}
		if !strings.Contains(string(doc.PreTOC), "<h2>Intro</h2>") {
			t.Errorf("Intro section should be in PreTOC: %s", doc.PreTOC)
		}
		if !strings.Contains(string(doc.Body), "<h2>Main</h2>") {
			t.Error("Main section should remain in body")
		}
	})

	t.Run("unknown heading is silently ignored", func(t *testing.T) {
		doc := &document.Document{Body: template.HTML(`<h2>Only</h2><p>text</p>`)}
		ExtractPreTOC(doc, []string{"NoSuchSection"})
		if string(doc.PreTOC) != "" {
			t.Errorf("PreTOC should be empty: %s", doc.PreTOC)
		}
		if !strings.Contains(string(doc.Body), "<h2>Only</h2>") {
			t.Error("body should be unchanged")
		}
	})
}
