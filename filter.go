package template

import (
	"reflect"
	"strings"
	"unicode"
	"utf8"
)

type filterFunc func(in Value, c *Context, arg Value) Value

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

func addFilter(in Value, c *Context, arg Value) Value {
	l := in.Int(c)
	r := arg.Int(c)
	return intValue(l + r)
}

func addslashesFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
	return stringValue(strings.Replace(str, "'", "\\'", -1))
}

func capfirstFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
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

func centerFilter(in Value, c *Context, arg Value) Value {
	count := int(arg.Int(c))
	if count <= 0 {
		return in
	}
	str := in.String(c)
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

func cutFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
	ch := arg.String(c)
	return stringValue(strings.Replace(str, ch, "", -1))
}

func dateFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func defaultFilter(in Value, c *Context, arg Value) Value {
	if b := in.Bool(c); b {
		return in
	}
	return arg
}

func defaultIfNilFilter(in Value, c *Context, arg Value) Value {
	if _, ok := in.(nilValue); !ok {
		return in
	}
	return arg
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortreversedFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func divisiblebyFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

// TODO: This isn't quite right; this filter should work anywhere in a filter chain so it probably needs to be treated specially
func escapeFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return stringValue(str)
}

func escapejsFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func filesizeformatFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

// Works on slices or arrays
func firstFilter(in Value, c *Context, arg Value) Value {
	in = in.Eval(c)
	v := in.Reflect(c)
	switch v.Kind() {
	case reflect.String:
		str := in.String(c)
		_, w := utf8.DecodeRuneInString(str)
		return stringValue(str[:w])
	case reflect.Array, reflect.Slice:
		return refToVal(v.Index(0))
	}
	return in
}

func fixAmpersandsFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
	return stringValue(strings.Replace(str, "&", "&amp;", -1))
}

func floatformatFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func forceEscapeFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return stringValue(str)
}

func getDigitFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func iriencodeFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func joinFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

// Works on slices or arrays
func lastFilter(in Value, c *Context, arg Value) Value {
	// TODO: Work on strings also
	list := in.Reflect(c)
	if list.Kind() != reflect.Array && list.Kind() != reflect.Slice {
		// TODO: Is this right?
		return in
	}
	len := list.Len()
	return refToVal(list.Index(len - 1))
}

func lengthFilter(in Value, c *Context, arg Value) Value {
	v := in.Reflect(c)
	switch v.Kind() {
	case reflect.String:
		return intValue(v.Len())
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return intValue(v.Len())
	}
	return in
}

func lengthIsFilter(in Value, c *Context, arg Value) Value {
	in = in.Eval(c)
	l := lengthFilter(in, c, nilValue(0)).Int(c)
	return boolValue(l == arg.Int(c))
}

func linebreaksFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "\n\n", "</p>", -1)
	str = strings.Replace(str, "\n", "<br />", -1)
	return stringValue(str)
}

func linebreaksbrFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
	// TODO: We can probably get better performance by implementing this ourselves
	return stringValue(strings.Replace(str, "\n", "<br />", -1))
}

func linenumbersFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func ljustFilter(in Value, c *Context, arg Value) Value {
	count := int(arg.Int(c))
	if count <= 0 {
		return in
	}
	str := in.String(c)
	runes := []int(str)
	if len(runes) >= count {
		return in
	}
	count -= len(runes)
	return stringValue(str + strings.Repeat(" ", count))
}

func lowerFilter(in Value, c *Context, arg Value) Value {
	str := in.String(c)
	return stringValue(strings.ToLower(str))
}

func makeListFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func phone2numericFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func pluralizeFilter(in Value, c *Context, arg Value) Value {
	var single string
	var plural string
	suffix := arg.String(c)
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

	if in.Uint(c) > 0 {
		return stringValue(plural)
	}
	return stringValue(single)
}

func pprintFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func randomFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func removeTagsFilter(in Value, c *Context, arg Value) Value {
	// TODO: Implement
	return in
}

func rjustFilter(in Value, c *Context, arg Value) Value {
	count := int(arg.Int(c))
	if count <= 0 {
		return in
	}
	str := in.String(c)
	runes := []int(str)
	if len(runes) >= count {
		return in
	}
	count -= len(runes)
	return stringValue(strings.Repeat(" ", count) + str)
}
