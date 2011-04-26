package template

import (
	"bytes"
	"testing"
)

// to save typing
type c map[string]interface{}

type templateTest struct {
	template string
	vars     map[string]interface{}
	out      string
}

type testStruct struct {
	a int
	b float64
}

var templateTests = []templateTest{
	{"hello", nil, "hello"},
	{"hello{", nil, "hello{"},
	{"hello{i", nil, "hello{i"},
	{"{# it's a comment #}", nil, ""},
	{"{{ 1 }}", nil, "1"},
	{"{{ 3.14 }}", nil, "3.14"},
	{"{{ 'hello' }}", nil, "hello"},
	{"{{ \"hello\" }}", nil, "hello"},
	{"{{ 1 }} {{ 2 }}", nil, "1 2"},
	{"{{ var }}", c{"var": "hello"}, "hello"},
	{" {{ var }}", c{"var": []int{1, 2, 3}}, " [1, 2, 3]"},
	{"{{ var }}", c{"var": map[int]string{1: "one"}}, "{1: 'one'}"},
	{"{{ var }}", c{"var": 2 + 2i}, "2+2i"},
	{"{{ 'hello'.1 }}", nil, "e"},
	{"{{ var.1 }}", c{"var": "hello"}, "e"},
	{"{{ var.0 }}", c{"var": []int{14}}, "14"},
	{"{{ var.13 }}", c{"var": [14]int{13: 11}}, "11"},
	{"{{ var.test }}", c{"var": map[string]string{"test": "hello"}}, "hello"},
	{"{{ var.42 }}", c{"var": map[int]int{42: 67}}, "67"},
	{"{{ var.42 }}", c{"var": map[int16]int16{42: 67}}, "67"},
	{"{{ var.a }}", c{"var": testStruct{4, 3.14}}, "4"},
	{"{{ var.b }}", c{"var": &testStruct{4, 3.14}}, "3.14"},
}

func testTemplates(t *testing.T, templates []templateTest) {
	for i, test := range templates {
		temp, err := ParseString(test.template)
		if err != nil {
			t.Errorf("#%d failed to parse: %s", i, err.String())
		}
		buf := bytes.NewBuffer(nil)
		temp.Execute(buf, test.vars)
		if buf.String() != test.out {
			t.Errorf("#%d got %q want %q", i, buf.String(), test.out)
		}
	}
}

func TestTemplate(t *testing.T) {
	testTemplates(t, templateTests)
}

// Benchmark taken from here: http://code.google.com/p/spitfire/source/browse/trunk/tests/perf/bigtable.py
var bench = `<table>
{% for row in table %}
<tr>{% for col in row %}{{ col|escape }}{% endfor %}</tr>
{% endfor %}
</table>
`

func BenchmarkTemplateParse(b *testing.B) {
	b.StopTimer()
	src := []byte(bench)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Parse(src)
	}
}

func BenchmarkTemplateParseString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseString(bench)
	}
}

func BenchmarkTemplateExecute(b *testing.B) {
	b.StopTimer()
	table := make([]map[string]int, 1000)
	for i := 0; i < 1000; i++ {
		table[i] = map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8, "i": 9, "j": 10}
	}
	t, _ := ParseString(bench)
	vars := c{"table": table}
	w := nilWriter(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		t.Execute(w, vars)
	}
}
