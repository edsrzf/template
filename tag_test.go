package template

import (
	"testing"
)

var tagTests = []templateTest{
	// Test firstof
	templateTest{"{% firstof %}", nil, ""},
	templateTest{"{% firstof var %}", nil, ""},
	templateTest{"{% firstof var %}", Context{"var": 1}, "1"},
	templateTest{"{% firstof var1 var2 %}", Context{"var1": nil, "var2": "y"}, "y"},
	templateTest{"{% firstof var1 var2 'default' %}", nil, "default"},

	// Test if
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
