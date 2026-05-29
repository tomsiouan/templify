package parser

import (
	"html/template"
	"regexp"
	"strings"

	"github.com/tomsiouan/templify/pkg/document"
)

var (
	reSectionHeading = regexp.MustCompile(`<h2[^>]*>([^<]*)</h2>`)
	reTableBlock     = regexp.MustCompile(`(?s)<table>.*?</table>`)
	reTableHead      = regexp.MustCompile(`(?s)<thead>(.*?)</thead>`)
	reTableBody      = regexp.MustCompile(`(?s)<tbody>(.*?)</tbody>`)
	reTableRow       = regexp.MustCompile(`(?s)<tr>(.*?)</tr>`)
	reTableCell      = regexp.MustCompile(`(?s)<t[dh][^>]*>(.*?)</t[dh]>`)
	reStripTags      = regexp.MustCompile(`<[^>]+>`)
)

// ExtractSections splits doc.Body by h2 headings and populates doc.Sections.
// Each section contains the rendered HTML between two consecutive h2s,
// and a parsed Table if a Markdown table is present.
func ExtractSections(doc *document.Document) {
	body := string(doc.Body)
	matches := reSectionHeading.FindAllStringIndex(body, -1)
	if len(matches) == 0 {
		return
	}

	sections := make(map[string]document.Section, len(matches))
	for i, loc := range matches {
		heading := reSectionHeading.FindStringSubmatch(body[loc[0]:loc[1]])[1]
		heading = strings.TrimSpace(heading)

		start := loc[1]
		end := len(body)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		content := strings.TrimSpace(body[start:end])

		sections[heading] = document.Section{
			HTML:  template.HTML(content),
			Table: parseTable(content),
		}
	}
	doc.Sections = sections
}

// parseTable extracts a two-dimensional slice of cell strings from the first
// HTML table in html. Row 0 contains header cells when a <thead> is present.
func parseTable(html string) [][]string {
	tbl := reTableBlock.FindString(html)
	if tbl == "" {
		return nil
	}

	var rows [][]string

	if head := reTableHead.FindStringSubmatch(tbl); head != nil {
		if row := parseCells(head[1]); len(row) > 0 {
			rows = append(rows, row)
		}
	}
	if body := reTableBody.FindStringSubmatch(tbl); body != nil {
		for _, tr := range reTableRow.FindAllStringSubmatch(body[1], -1) {
			if row := parseCells(tr[1]); len(row) > 0 {
				rows = append(rows, row)
			}
		}
	}
	return rows
}

// parseCells returns the trimmed text content of each <td> or <th> in rowHTML,
// stripping any inner HTML tags.
func parseCells(rowHTML string) []string {
	cells := reTableCell.FindAllStringSubmatch(rowHTML, -1)
	if len(cells) == 0 {
		return nil
	}
	row := make([]string, len(cells))
	for i, c := range cells {
		row[i] = strings.TrimSpace(reStripTags.ReplaceAllString(c[1], ""))
	}
	return row
}
