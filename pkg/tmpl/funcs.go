package tmpl

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

// templateFuncs returns the FuncMap exposed to all HTML templates, providing
// currency formatting, arithmetic helpers, and table-cell accessors.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"toFloat": toFloat,
		"mul":     func(a, b float64) float64 { return a * b },
		"add":     func(a, b float64) float64 { return a + b },
		"sub":     func(a, b float64) float64 { return a - b },
		"pct":     func(base, rate float64) float64 { return base * rate / 100 },
		"sumCol":     sumCol,
		"sumProduct": sumProduct,
		"currency": func(f float64) string {
			s := fmt.Sprintf("%.2f", f)
			parts := strings.Split(s, ".")
			intPart := parts[0]
			var out []byte
			for i, c := range intPart {
				if i > 0 && (len(intPart)-i)%3 == 0 {
					out = append(out, ' ')
				}
				out = append(out, byte(c))
			}
			return string(out) + "," + parts[1] + " €"
		},
		"rowSlice": func(rows [][]string, from int) [][]string {
			if from >= len(rows) {
				return nil
			}
			return rows[from:]
		},
		"cell": func(row []string, i int) string {
			if i < 0 || i >= len(row) {
				return ""
			}
			return row[i]
		},
		"lastCell": func(row []string) string {
			if len(row) == 0 {
				return ""
			}
			return row[len(row)-1]
		},
		"prevCell": func(row []string) string {
			if len(row) < 2 {
				return ""
			}
			return row[len(row)-2]
		},
		"initCells": func(row []string) []string {
			if len(row) < 2 {
				return nil
			}
			return row[:len(row)-2]
		},
		"sumProductLast": sumProductLast,
	}
}

// toFloat parses a currency-like string to float64, accepting spaces as
// thousands separators, a comma as the decimal mark, and trailing €/$/% symbols.
func toFloat(s string) float64 {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimRight(s, "€$%")
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

// sumProductLast multiplies the last two columns of each row and sums the results.
// Convention: second-to-last = quantity, last = unit price.
func sumProductLast(rows [][]string) float64 {
	var total float64
	for _, row := range rows {
		if len(row) >= 2 {
			total += toFloat(row[len(row)-2]) * toFloat(row[len(row)-1])
		}
	}
	return total
}

// sumProduct multiplies rows[i][colA] by rows[i][colB] for each row and returns the total.
func sumProduct(rows [][]string, colA, colB int) float64 {
	var total float64
	for _, row := range rows {
		if colA < len(row) && colB < len(row) {
			total += toFloat(row[colA]) * toFloat(row[colB])
		}
	}
	return total
}

// sumCol sums the values in column col across all rows.
func sumCol(rows [][]string, col int) float64 {
	var total float64
	for _, row := range rows {
		if col < len(row) {
			total += toFloat(row[col])
		}
	}
	return total
}
