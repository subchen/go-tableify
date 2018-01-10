package tableify

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	Table struct {
		headers    []string
		minwidths  []int
		realwidths []int
		rows       [][]string

		Margin    int
		SplitLine bool
		EmptyText string

		FormatFunc func(header string, value interface{}) string
	}
)

func New() *Table {
	return &Table{
		Margin:     3,
		SplitLine:  true,
		FormatFunc: format,
	}
}

func (t *Table) SetHeaders(headers ...string) {
	t.headers = headers
	t.minwidths = make([]int, len(headers))
	t.realwidths = make([]int, len(headers))
}

func (t *Table) SetWidths(widths ...int) {
	if len(t.headers) != len(widths) {
		panic("widths count is not matched with headers")
	}
	t.minwidths = widths
}

func (t *Table) AddRow(data ...interface{}) {
	if len(data) != len(t.headers) {
		panic("columns count is not matched with headers")
	}

	// format as []string
	row := make([]string, len(t.headers))
	for i, v := range data {
		row[i] = t.FormatFunc(t.headers[i], v)

		// calc real width
		if len(row[i]) > t.realwidths[i] {
			t.realwidths[i] = len(row[i])
		}
		if t.realwidths[i] < t.minwidths[i] {
			t.realwidths[i] = t.minwidths[i]
		}

	}

	t.rows = append(t.rows, row)
}

func (t *Table) AddRowList(rows ...[]string) {
	for _, row := range rows {
		t.AddRow(asInterfaces(row)...)
	}
}

func (t *Table) Print() {
	totalWidth := 0
	format := ""

	// confirm data.width > header.width
	for i, h := range t.headers {
		if len(h) > t.realwidths[i] {
			t.realwidths[i] = len(h)
		}
	}

	for i, w := range t.realwidths {
		margin := t.Margin
		if i == len(t.realwidths)-1 {
			// Don't add margin for the last column
			margin = 0
		}
		format += "%-" + strconv.FormatInt(int64(w), 10) + "s" + strings.Repeat(" ", margin)
		totalWidth += w + margin
	}
	format += "\n"

	// print headers
	fmt.Printf(format, asInterfaces(t.headers)...)

	// print split line
	if t.SplitLine {
		fmt.Println(strings.Repeat("-", int(totalWidth)))
	}

	// print rows
	for _, row := range t.rows {
		fmt.Printf(format, asInterfaces(row)...)
	}

	// print empty text if no rows
	if len(t.rows) == 0 && len(t.EmptyText) > 0 {
		fmt.Println(t.EmptyText)
	}
}

func format(header string, value interface{}) string {
	if value == nil {
		return ""
	}

	// fmt.Stringer -> string
	if stringer, ok := value.(fmt.Stringer); ok {
		return stringer.String()
	}

	// []byte -> string
	if bytes, ok := value.([]byte); ok {
		return string(bytes)
	}

	// float64 -> string
	if f, ok := value.(float64); ok {
		return fmt.Sprintf("%.4f", f)
	}

	// float32 -> string
	if f, ok := value.(float32); ok {
		return fmt.Sprintf("%.2f", f)
	}

	// * -> string
	return fmt.Sprintf("%v", value)
}

func asInterfaces(list []string) []interface{} {
	vals := make([]interface{}, len(list))
	for i, v := range list {
		vals[i] = v
	}
	return vals
}
