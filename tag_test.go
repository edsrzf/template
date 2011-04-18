package template

import (
	"testing"
)

var tagTests = []templateTest{
	// firstof
	templateTest{"{% firstof %}", nil, ""},
	templateTest{"{% firstof var %}", nil, ""},
	templateTest{"{% firstof var %}", Context{"var": 1}, "1"},
	templateTest{"{% firstof var1 var2 %}", Context{"var1": nil, "var2": "y"}, "y"},
	templateTest{"{% firstof var1 var2 'default' %}", nil, "default"},

	// for
	templateTest{"{% for l in 'hello' %}{{l}}{% endfor %}", nil, "hello"},
	templateTest{"{% for l in 'hello' %}{{l}} {% endfor %}", nil, "h e l l o "},
	templateTest{"{% for l in '' %}{{l}}{% else %}hi{% endfor %}", nil, "hi"},
	templateTest{"{% for v in var %}{{v}}{% endfor %}", Context{"var": []int{1, 2, 3}}, "123"},
	templateTest{"{% for v in var %}{{v}} {% endfor %}", Context{"var": &testStruct{1, 3.14}}, "1 3.14 "},
	templateTest{"{% for v in var %}{{v}}{% endfor %} {{v}}", Context{"var": []int{1, 2, 3}, "v": "hi"}, "123 hi"},

	// if
	templateTest{"{% if 1 %}hi{% endif %}", nil, "hi"},
	templateTest{"{% if 0 %}hi{% endif %}", nil, ""},
	templateTest{"{% if 'hi' %}hi{% endif %}", nil, "hi"},
	templateTest{"{% if '' %}hi{% endif %}", nil, ""},
	templateTest{"{% if var %}hi{% endif %}", nil, ""},
	templateTest{"{% if var %}hi{% endif %}", Context{"var": 1}, "hi"},
}

func TestTags(t *testing.T) {
	testTemplates(t, tagTests)
}
