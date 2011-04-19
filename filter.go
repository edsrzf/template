package template

import (
	"reflect"
	"strings"
	"unicode"
	"utf8"
)

type filterFunc func(in Value, s Stack, arg Value) Value

type filter struct {
	f    filterFunc
	args Value
}

type argType int

const (
	NoArg  argType = iota // No argument allowed
	OptArg                // Optional argument
	ReqArg                // Argument required
)

type regFilter struct {
	f   filterFunc
	arg argType
}

var filters = map[string]*regFilter{
	"add":            &regFilter{addFilter, ReqArg},
	"addslashes":     &regFilter{addslashesFilter, NoArg},
	"capfirst":       &regFilter{capfirstFilter, NoArg},
	"center":         &regFilter{centerFilter, ReqArg},
	"cut":            &regFilter{cutFilter, ReqArg},
	"default":        &regFilter{defaultFilter, ReqArg},
	"default_if_nil": &regFilter{defaultIfNilFilter, ReqArg},
	"escape":         &regFilter{escapeFilter, NoArg},
	"first":          &regFilter{firstFilter, NoArg},
	"lower":          &regFilter{lowerFilter, NoArg},
}

func addFilter(in Value, s Stack, arg Value) Value {
	l := in.Int(s)
	r := arg.Int(s)
	return intValue(l + r)
}

func addslashesFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	return stringValue(strings.Replace(str, "'", "\\'", -1))
}

func capfirstFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	if len(str) == 0 {
		return in
	}
	b := []byte(str)
	rune, _ := utf8.DecodeRune(b)
	rune = unicode.ToUpper(rune)
	// This assumes that the upper case rune is the same width as the lower case rune.
	// It's almost always true (might even be always).
	utf8.EncodeRune(b, rune)
	return stringValue(b)
}

func centerFilter(in Value, s Stack, arg Value) Value {
	c := arg.Int(s)
	count := int(c)
	if count <= 0 {
		return in
	}
	str := in.String(s)
	runes := []int(str)
	l := len(runes)
	if l >= count {
		return in
	}
	count -= l
	half := count / 2
	count = count - half
	if count == half {
		spaces := strings.Repeat(" ", count)
		return stringValue(spaces + str + spaces)
	}
	return stringValue(strings.Repeat(" ", half) + str + strings.Repeat(" ", count))
}

// Our cutFilter is slightly more forgiving than Django's. It allows the argument to be an integer.
// {{ 123|cut:2 }} will output "13".
func cutFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	ch := arg.String(s)
	return stringValue(strings.Replace(str, ch, "", -1))
}

func dateFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func defaultFilter(in Value, s Stack, arg Value) Value {
	if b := in.Bool(s); b {
		return in
	}
	return arg
}

func defaultIfNilFilter(in Value, s Stack, arg Value) Value {
	if _, ok := in.(nilValue); !ok {
		return in
	}
	return arg
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortreversedFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func divisiblebyFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

// TODO: This isn't quite right; this filter should work anywhere in a filter chain so it probably needs to be treated specially
func escapeFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return stringValue(str)
}

func escapejsFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func filesizeformatFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

// Works on slices or arrays
func firstFilter(in Value, s Stack, arg Value) Value {
	in = in.Eval(s)
	v := in.Reflect(s)
	switch v.Kind() {
	case reflect.String:
		str := in.String(s)
		_, w := utf8.DecodeRuneInString(str)
		return stringValue(str[:w])
	case reflect.Array, reflect.Slice:
		return refToVal(v.Index(0))
	}
	return in
}

func fixAmpersandsFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	return stringValue(strings.Replace(str, "&", "&amp;", -1))
}

func floatformatFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func forceEscapeFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return stringValue(str)
}

func getDigitFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func iriencodeFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func joinFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

// Works on slices or arrays
func lastFilter(in Value, s Stack, arg Value) Value {
	// TODO: Work on strings also
	list := in.Reflect(s)
	if list.Kind() != reflect.Array && list.Kind() != reflect.Slice {
		// TODO: Is this right?
		return in
	}
	len := list.Len()
	return refToVal(list.Index(len - 1))
}

func lengthFilter(in Value, s Stack, arg Value) Value {
	v := in.Reflect(s)
	switch v.Kind() {
	case reflect.String:
		return intValue(v.Len())
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return intValue(v.Len())
	}
	return in
}

func lengthIsFilter(in Value, s Stack, arg Value) Value {
	in = in.Eval(s)
	l := lengthFilter(in, s, nilValue(0)).Int(s)
	return boolValue(l == arg.Int(s))
}

func linebreaksFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "\n\n", "</p>", -1)
	str = strings.Replace(str, "\n", "<br />", -1)
	return stringValue(str)
}

func linebreaksbrFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	// TODO: We can probably get better performance by implementing this ourselves
	return stringValue(strings.Replace(str, "\n", "<br />", -1))
}

func linenumbersFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func ljustFilter(in Value, s Stack, arg Value) Value {
	count := int(arg.Int(s))
	if count <= 0 {
		return in
	}
	str := in.String(s)
	runes := []int(str)
	if len(runes) >= count {
		return in
	}
	count -= len(runes)
	return stringValue(str + strings.Repeat(" ", count))
}

func lowerFilter(in Value, s Stack, arg Value) Value {
	str := in.String(s)
	return stringValue(strings.ToLower(str))
}

func makeListFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func phone2numericFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func pluralizeFilter(in Value, s Stack, arg Value) Value {
	var single string
	var plural string
	suffix := arg.String(s)
	if suffix == "" {
		plural = "s"
	} else {
		args := strings.Split(suffix, ",", 2)
		if len(args) == 1 {
			plural = args[0]
		} else {
			single = args[0]
			plural = args[1]
		}
	}

	if in.Uint(s) > 0 {
		return stringValue(plural)
	}
	return stringValue(single)
}

func pprintFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func randomFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func removeTagsFilter(in Value, s Stack, arg Value) Value {
	// TODO: Implement
	return in
}

func rjustFilter(in Value, s Stack, arg Value) Value {
	count := int(arg.Int(s))
	if count <= 0 {
		return in
	}
	str := in.String(s)
	runes := []int(str)
	if len(runes) >= count {
		return in
	}
	count -= len(runes)
	return stringValue(strings.Repeat(" ", count) + str)
}
