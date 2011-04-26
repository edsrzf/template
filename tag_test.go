package template

import (
	"testing"
)

var parentTemplate = MustParseString("parent start {% block title %}parent title{% endblock %}\n")
var varTemplate = MustParseString("{{ var }}")

var tagTests = []templateTest{
	// cycle
	{"{% for c in 'abcd' %}{% cycle 1 'a' var %}{% endfor %}", c{"var": 3.14}, "1a3.141"},
	{"{% for c in 'abcd' %}{% cycle 1 2 3 %}{% endfor %}{% for c in 'ab' %}{% cycle 1 2 %}{% endfor %}", nil, "123112"},

	// firstof
	{"{% firstof %}", nil, ""},
	{"{% firstof var %}", nil, ""},
	{"{% firstof var %}", c{"var": 1}, "1"},
	{"{% firstof var1 var2 %}", c{"var1": nil, "var2": "y"}, "y"},
	{"{% firstof var1 var2 'default' %}", nil, "default"},

	// for
	{"{% for l in 'hello' %}{{l}}{% endfor %}", nil, "hello"},
	{"{% for l in 'hello' %}{{l}} {% endfor %}", nil, "h e l l o "},
	{"{% for l in '' %}{{l}}{% else %}hi{% endfor %}", nil, "hi"},
	{"{% for v in var %}{{v}}{% endfor %}", c{"var": []int{1, 2, 3}}, "123"},
	{"{% for v in var %}{{v}} {% endfor %}", c{"var": &testStruct{1, 3.14}}, "1 3.14 "},
	{"{% for v in var %}{{v}}{% endfor %} {{v}}", c{"var": []int{1, 2, 3}, "v": "hi"}, "123 hi"},

	// if
	{"{% if 1 %}hi{% endif %}", nil, "hi"},
	{"{% if 0 %}hi{% endif %}", nil, ""},
	{"{% if 'hi' %}hi{% endif %}", nil, "hi"},
	{"{% if '' %}hi{% endif %}", nil, ""},
	{"{% if var %}hi{% endif %}", nil, ""},
	{"{% if var %}hi{% endif %}", c{"var": 1}, "hi"},

	// ifchanged
	{"{% for n in '122344' %}{% ifchanged n %}c{% endifchanged %}{% endfor %}", nil, "cccc"},
	{"{% for n in '122344' %}{% ifchanged n %}c{% else %}s{% endifchanged %}{% endfor %}", nil, "ccsccs"},
	{"{% ifchanged var %}c{% endifchanged %}", nil, "c"},

	// include
	{"{% include t %}", c{"t": parentTemplate}, "parent start parent title\n"},
	{"{% include t %}", c{"t": varTemplate, "var": "hello"}, "hello"},

	// set
	{"{% set var1 1 %}{% set var2 'hi' %}{% set var3 3.14 %}{{ var1 }} {{ var2 }} {{ var3 }}", nil, "1 hi 3.14"},
	{"{% set setvar var %}{{ setvar }}", c{"var": "test"}, "test"},
	{"{% set var2 var1 %}{% set var1 2 %}{{ var1 }} {{ var2 }}", c{"var1": 1}, "2 1"},

	// with
	{"{% with %}{% set var 1 %}{{ var }}{% endwith %} {{ var }}", nil, "1 "},

	// inheritance
	{"{% extends parent %}{% override title %}child title{% endoverride %}{% endextends %}", c{"parent": parentTemplate}, "parent start child title\n"},
	{"{% extends parent %}{% override title %}child title{% endoverride %}{% endextends %}", c{"parent": "testdata/parent"}, "parent start child title\n"},
}

func TestTags(t *testing.T) {
	testTemplates(t, tagTests)
}
