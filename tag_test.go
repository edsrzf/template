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
}

func TestTags(t *testing.T) {
	testTemplates(t, tagTests)
}
