package template

import (
	"testing"
)

var filterTests = []templateTest{
	// Test add
	templateTest{"{{ var|add:1 }}", Context{"var": 14}, "15"},
	templateTest{"{{ var|add:-1 }}", Context{"var": 14}, "13"},
	templateTest{"{{ var|add:\"1\" }}", Context{"var": 14}, "15"},
	templateTest{"{{ var|add:'-1' }}", Context{"var": 14}, "13"},
	templateTest{"{{ var|add:1 }}", Context{"var": "14"}, "15"},
	templateTest{"{{ var|add:1 }}", nil, ""},
	templateTest{"{{ var|add:1 }}", Context{"var": "a"}, "a"},

	// Test addslashes
	templateTest{"{{ var|addslashes }}", Context{"var": "I'm using Django"}, "I\\'m using Django"},

	// Test capfirst
	templateTest{"{{ var1|capfirst }} {{ var2|capfirst }} {{ var3|capfirst }} {{ var4|capfirst }}",
		Context{"var1": "var", "var2": "Var", "var3": 13}, "Var Var 13 "},

	// Test center
	templateTest{"{{ var1|center:2 }} {{ var2|center:3 }} {{ var3|center:3 }}",
		Context{"var1": "a", "var2": 1}, "a   1     "},

	// Test cut
	templateTest{"{{ var1|cut:2 }} {{ var2|cut:\"2\" }} {{ var3|cut:2 }} {{ var4|cut:'2' }} {{ var5|cut:' ' }}",
		Context{"var1": 123, "var2": "123", "var3": "123", "var4": 123}, "13 13 13 13 "},

	// Test default
	templateTest{"{{ var1|default:'def' }} {{ var2|default:'def' }} {{ var3|default:'def' }}",
		Context{"var1": 14, "var2": nil}, "14 def def"},

	// Test default_if_none
	templateTest{"{{ var1|default_if_nil:'def' }} {{ var2|default_if_nil:'def' }} {{ var3|default_if_nil:'def' }}",
		Context{"var1": 14, "var2": nil}, "14 def def"},

	// Test first
	templateTest{"{{ 'hello'|first }} {{ var1|first }} {{ var2|first }} {{ var3|first }}",
		Context{"var1": []int{1, 2, 3}, "var2": [...]byte{4, 5, 6}}, "h 1 4 "},

	// Test lower
	templateTest{"{{ var|lower }}", Context{"var": "VaR"}, "var"},
}

func TestFilters(t *testing.T) {
	testTemplates(t, filterTests)
}
