package template

import (
	"testing"
)

var tagTests = []templateTest{
	// cycle
	templateTest{"{% for c in 'abcd' %}{% cycle 1 'a' var %}{% endfor %}", Context{"var": 3.14}, "1a3.141"},

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

	// ifchanged
	templateTest{"{% for n in '122344' %}{% ifchanged n %}c{% endifchanged %}{% endfor %}", nil, "cccc"},
	templateTest{"{% for n in '122344' %}{% ifchanged n %}c{% else %}s{% endifchanged %}{% endfor %}", nil, "ccsccs"},

	// set
	templateTest{"{% set var1 1 %}{% set var2 'hi' %}{% set var3 3.14 %}{{ var1 }} {{ var2 }} {{ var3 }}", nil, "1 hi 3.14"},
	templateTest{"{% set setvar var %}{{ setvar }}", Context{"var": "test"}, "test"},
	templateTest{"{% set var2 var1 %}{% set var1 2 %}{{ var1 }} {{ var2 }}", Context{"var1": 1}, "2 1"},

	// with
	templateTest{"{% with %}{% set var 1 %}{{ var }}{% endwith %} {{ var }}", nil, "1 "},
}

func TestTags(t *testing.T) {
	testTemplates(t, tagTests)
}
