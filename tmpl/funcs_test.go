package tmpl

import (
	"testing"

)

func TestToFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"1000", 1000},
		{"1 000,50", 1000.50},
		{"1.50€", 1.50},
		{"42$", 42},
		{"50%", 50},
		{"", 0},
		{"abc", 0},
		{"1 234 567,89", 1234567.89},
		{"0,00 €", 0},
	}
	for _, tc := range tests {
		if got := toFloat(tc.input); got != tc.want {
			t.Errorf("toFloat(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestSumCol(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		col  int
		want float64
	}{
		{"empty rows", nil, 0, 0},
		{"single column", [][]string{{"1"}, {"2"}, {"3"}}, 0, 6},
		{"second column", [][]string{{"a", "10"}, {"b", "20"}}, 1, 30},
		{"out of bounds row skipped", [][]string{{"1"}, {"2", "5"}}, 1, 5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sumCol(tc.rows, tc.col); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSumProduct(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		colA int
		colB int
		want float64
	}{
		{"empty rows", nil, 0, 1, 0},
		{"two rows", [][]string{{"2", "3"}, {"4", "5"}}, 0, 1, 26},
		{"row missing colB skipped", [][]string{{"2"}, {"3", "4"}}, 0, 1, 12},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sumProduct(tc.rows, tc.colA, tc.colB); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSumProductLast(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		want float64
	}{
		{"empty rows", nil, 0},
		{"single column row skipped", [][]string{{"5"}}, 0},
		{"two columns", [][]string{{"2", "3"}, {"4", "5"}}, 26},
		{"three columns uses last two", [][]string{{"label", "2", "10"}}, 20},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sumProductLast(tc.rows); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMul(t *testing.T) {
	mul := templateFuncs()["mul"].(func(float64, float64) float64)
	if got := mul(3, 4); got != 12 {
		t.Errorf("mul(3, 4) = %v, want 12", got)
	}
}

func TestAdd(t *testing.T) {
	add := templateFuncs()["add"].(func(float64, float64) float64)
	if got := add(2.5, 1.5); got != 4 {
		t.Errorf("add(2.5, 1.5) = %v, want 4", got)
	}
}

func TestSub(t *testing.T) {
	sub := templateFuncs()["sub"].(func(float64, float64) float64)
	if got := sub(5, 3); got != 2 {
		t.Errorf("sub(5, 3) = %v, want 2", got)
	}
}

func TestPct(t *testing.T) {
	pct := templateFuncs()["pct"].(func(float64, float64) float64)
	if got := pct(200, 10); got != 20 {
		t.Errorf("pct(200, 10) = %v, want 20", got)
	}
}

func TestRowSlice(t *testing.T) {
	rowSlice := templateFuncs()["rowSlice"].(func([][]string, int) [][]string)
	rows := [][]string{{"a"}, {"b"}, {"c"}}
	tests := []struct {
		from int
		want int
	}{
		{0, 3},
		{1, 2},
		{3, 0},
		{10, 0},
	}
	for _, tc := range tests {
		got := rowSlice(rows, tc.from)
		if len(got) != tc.want {
			t.Errorf("rowSlice(rows, %d) len = %d, want %d", tc.from, len(got), tc.want)
		}
	}
}

func TestCell(t *testing.T) {
	cell := templateFuncs()["cell"].(func([]string, int) string)
	row := []string{"a", "b", "c"}
	if got := cell(row, 1); got != "b" {
		t.Errorf("cell(row, 1) = %q, want b", got)
	}
	if got := cell(row, -1); got != "" {
		t.Errorf("cell(row, -1) = %q, want empty", got)
	}
	if got := cell(row, 10); got != "" {
		t.Errorf("cell(row, 10) = %q, want empty", got)
	}
}

func TestLastCell(t *testing.T) {
	lastCell := templateFuncs()["lastCell"].(func([]string) string)
	if got := lastCell([]string{"a", "b", "c"}); got != "c" {
		t.Errorf("got %q, want c", got)
	}
	if got := lastCell(nil); got != "" {
		t.Errorf("lastCell(nil) = %q, want empty", got)
	}
}

func TestPrevCell(t *testing.T) {
	prevCell := templateFuncs()["prevCell"].(func([]string) string)
	if got := prevCell([]string{"a", "b", "c"}); got != "b" {
		t.Errorf("got %q, want b", got)
	}
	if got := prevCell([]string{"a"}); got != "" {
		t.Errorf("prevCell with 1 element = %q, want empty", got)
	}
}

func TestInitCells(t *testing.T) {
	initCells := templateFuncs()["initCells"].(func([]string) []string)
	if got := initCells([]string{"a", "b", "c", "d"}); len(got) != 2 || got[0] != "a" {
		t.Errorf("got %v, want [a b]", got)
	}
	if got := initCells([]string{"a"}); got != nil {
		t.Errorf("initCells with 1 element = %v, want nil", got)
	}
}

func TestCurrency(t *testing.T) {
	currency := templateFuncs()["currency"].(func(float64) string)

	tests := []struct {
		input float64
		want  string
	}{
		{0, "0,00 €"},
		{1.5, "1,50 €"},
		{1000, "1 000,00 €"},
		{1000.50, "1 000,50 €"},
		{1234567.89, "1 234 567,89 €"},
	}
	for _, tc := range tests {
		if got := currency(tc.input); got != tc.want {
			t.Errorf("currency(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
