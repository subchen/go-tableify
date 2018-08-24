package tableify

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type (
	Table struct {
		headers []string
		rows    [][]string

		minwidths []int
		formats   []string

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
	t.formats = make([]string, len(headers))
}

// SetStructHeaders to add headers from struct tags
// Example struct {
//      Name        string  `tableify:"-"`         // header is fieldname: Name
//      Files       int     `tableify:"-"`
//      Age         float64 `tableify:"-,0,%.2f"`  // width is 0, format is "%.2f"
//      LastUpdated string  `tableify:"Updated"`   // header is "Updated"
//      Desc        string                         // not as header
//
func (t *Table) SetHeadersFromStruct(obj interface{}) {
	obj = indirect(obj)
	rtype := reflect.TypeOf(obj)
	if rtype.Kind() != reflect.Struct {
		panic("given type is not a struct")
	}

	// reset
	t.headers = nil
	t.minwidths = nil
	t.formats = nil

	var header string
	var width int
	var format string
	var err error

	// revieve headers from strcut
	numFields := rtype.NumField()
	for i := 0; i < numFields; i++ {
		field := rtype.Field(i)
		tag := field.Tag.Get("tableify")
		if tag != "" {
			if strings.Contains(tag, ",") {
				parts := strings.Split(tag, ",")
				// 0: header
				header = parts[0]

				// 1: width
				if parts[1] == "" {
					width = 0
				} else {
					width, err = strconv.Atoi(parts[1])
					if err != nil{
						panic(fmt.Errorf("width is not an integer: %s", parts[1], err.Error()))
					}
				}

				// 2: format
				if len(parts) > 2 {
					format = parts[2]
				} else {
					format = ""
				}
			} else {
				header = tag
				width = 0
				format = ""
			}

			if header == "" || header == "-" {
				header = field.Name
			}

			t.headers = append(t.headers, header)
			t.minwidths = append(t.minwidths, width)
			t.formats = append(t.formats, format)
		}
	}
}

func (t *Table) SetWidths(widths ...int) {
	if len(t.headers) != len(widths) {
		panic("widths count is not matched with headers")
	}
	t.minwidths = widths
}

func (t *Table) SetFormats(formats ...string) {
	if len(t.headers) != len(formats) {
		panic("formats count is not matched with headers")
	}
	t.formats = formats
}

func (t *Table) AddRow(values ...interface{}) {
	if len(values) != len(t.headers) {
		panic("columns count is not matched with headers")
	}

	// format as []string
	row := make([]string, len(t.headers))
	for i, v := range values {
		if t.formats[i] != "" {
			row[i] = fmt.Sprintf(t.formats[i], v)
		} else {
			row[i] = t.FormatFunc(t.headers[i], v)
		}
	}

	t.rows = append(t.rows, row)
}

func (t *Table) AddRowList(rows ...[]string) {
	for _, row := range rows {
		t.AddRow(asInterfaces(row)...)
	}
}

func (t *Table) AddRowObject(obj interface{}) {
	obj = indirect(obj)
	rtype := reflect.TypeOf(obj)
	rvalue := reflect.ValueOf(obj)
	if rtype.Kind() != reflect.Struct {
		panic("given type is not a struct")
	}

	var values []interface{}

	// revieve headers from strcut
	numFields := rtype.NumField()
	for i := 0; i < numFields; i++ {
		field := rtype.Field(i)
		tag := field.Tag.Get("tableify")
		if tag != "" {
			v := rvalue.Field(i).Interface()
			values = append(values, v)
		}
	}
	t.AddRow(values...)
}

func (t *Table) AddRowObjectList(objs interface{}) {
	rtype := reflect.TypeOf(objs)
	if rtype.Kind() != reflect.Slice && rtype.Kind() != reflect.Array {
		panic("given type is not slice or array")
	}

	rvalue := reflect.ValueOf(objs)
	length := rvalue.Len()
	for i := 0; i < length; i++ {
		elem := rvalue.Index(i)
		t.AddRowObject(elem.Interface())
	}
}

func (t *Table) Print() {
	fullwidth := 0
	fullformat := ""

	realwidths := make([]int, len(t.headers))

	// calc realwidths for each column
	for _, row := range t.rows {
		for i, v := range row {
			if realwidths[i] < len(v) {
				realwidths[i] = len(v)
			}
			if realwidths[i] < len(t.headers[i]) {
				realwidths[i] = len(t.headers[i])
			}
			if realwidths[i] < t.minwidths[i] {
				realwidths[i] = t.minwidths[i]
			}
		}
	}

	for i, w := range realwidths {
		margin := t.Margin
		if i == len(realwidths)-1 {
			margin = 0 // no margin for last column
		}
		fullformat += "%-" + strconv.FormatInt(int64(w), 10) + "s" + strings.Repeat(" ", margin)
		fullwidth += w + margin
	}
	fullformat += "\n"

	// print headers
	fmt.Printf(fullformat, asInterfaces(t.headers)...)

	// print split line
	if t.SplitLine {
		fmt.Println(strings.Repeat("-", int(fullwidth)))
	}

	// print rows
	for _, row := range t.rows {
		fmt.Printf(fullformat, asInterfaces(row)...)
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

// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
