package template

import (
	"bytes"
	"os"
	"testing"
)

type templateTest struct {
	template string
	context  Context
	out      string
}

type testStruct struct {
	a int
	b float64
}

var templateTests = []templateTest{
	templateTest{"hello", nil, "hello"},
	templateTest{"hello{", nil, "hello{"},
	templateTest{"hello{i", nil, "hello{i"},
	templateTest{"{# it's a comment #}", nil, ""},
	templateTest{"{{ 1 }}", nil, "1"},
	templateTest{"{{ 3.14 }}", nil, "3.14"},
	templateTest{"{{ 'hello' }}", nil, "hello"},
	templateTest{"{{ \"hello\" }}", nil, "hello"},
	templateTest{"{{ 1 }} {{ 2 }}", nil, "1 2"},
	templateTest{"{{ var }}", Context{"var": "hello"}, "hello"},
	templateTest{" {{ var }}", Context{"var": []int{1, 2, 3}}, " [1, 2, 3]"},
	templateTest{"{{ var }}", Context{"var": map[int]string{1: "one"}}, "{1: 'one'}"},
	templateTest{"{{ var.1 }}", Context{"var": "hello"}, "e"},
	templateTest{"{{ var.0 }}", Context{"var": []int{14}}, "14"},
	templateTest{"{{ var.13 }}", Context{"var": [14]int{13: 11}}, "11"},
	templateTest{"{{ var.test }}", Context{"var": map[string]string{"test": "hello"}}, "hello"},
	templateTest{"{{ var.42 }}", Context{"var": map[int]int{42: 67}}, "67"},
	templateTest{"{{ var.42 }}", Context{"var": map[int16]int16{42: 67}}, "67"},
	templateTest{"{{ var.a }}", Context{"var": testStruct{4, 3.14}}, "4"},
	templateTest{"{{ var.b }}", Context{"var": &testStruct{4, 3.14}}, "3.14"},
}

func testTemplates(t *testing.T, templates []templateTest) {
	for i, test := range templates {
		temp, err := ParseString(test.template)
		if err != nil {
			t.Errorf("#%d failed to parse: %s", i, err.String())
		}
		buf := bytes.NewBuffer(nil)
		temp.Execute(buf, test.context)
		if buf.String() != test.out {
			t.Errorf("#%d got %s want %s", i, buf.String(), test.out)
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

type nilWriter int

func (w nilWriter) Write(b []byte) (int, os.Error) {
	return len(b), nil
}

func BenchmarkTemplateExecute(b *testing.B) {
	b.StopTimer()
	table := make([]map[string]int, 1000)
	for i := 0; i < 1000; i++ {
		table[i] = map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8, "i": 9, "j": 10}
	}
	t, _ := ParseString(bench)
	c := Context{"table": table}
	w := nilWriter(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		t.Execute(w, c)
	}
}
