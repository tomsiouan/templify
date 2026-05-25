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
