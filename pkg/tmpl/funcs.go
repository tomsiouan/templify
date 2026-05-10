package tmpl

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"toFloat": toFloat,
		"mul":     func(a, b float64) float64 { return a * b },
		"add":     func(a, b float64) float64 { return a + b },
		"sub":     func(a, b float64) float64 { return a - b },
		"pct":     func(base, rate float64) float64 { return base * rate / 100 },
		"sumCol":  sumCol,
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
	}
}

func toFloat(s string) float64 {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimRight(s, "€$%")
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func sumCol(rows [][]string, col int) float64 {
	var total float64
	for _, row := range rows {
		if col < len(row) {
			total += toFloat(row[col])
		}
	}
	return total
}
