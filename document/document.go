package document

import "html/template"

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

// Section is one h2-delimited block of a document body.
// Table holds the parsed rows of the first table found in the section
// (row 0 = headers, subsequent rows = data). Nil if no table is present.
type Section struct {
	HTML  template.HTML
	Table [][]string
}

type Document struct {
	Title   string
	Author  string
	Date    string
	Body    template.HTML
	PreTOC  template.HTML
	Meta    map[string]any
	TOC     []TocEntry
	Figures []FigureEntry

	// Sections is the body split by h2 headings, keyed by heading text.
	// Available for bundle templates that use semantic Markdown.
	Sections map[string]Section
}
