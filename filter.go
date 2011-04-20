package template

import (
	"reflect"
	"strings"
	"unicode"
	"utf8"
)

type filterFunc func(in Expr, c *Context, arg Expr) Value

type filter struct {
	f    filterFunc
	args Expr
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

func addFilter(in Expr, c *Context, arg Expr) Value {
	l := in.Eval(c).Int()
	r := arg.Eval(c).Int()
	return intValue(l + r)
}

func addslashesFilter(in Expr, c *Context, arg Expr) Value {
	str := in.Eval(c).String()
	return stringValue(strings.Replace(str, "'", "\\'", -1))
}

func capfirstFilter(in Expr, c *Context, arg Expr) Value {
	inVal := in.Eval(c)
	str := inVal.String()
	if len(str) == 0 {
		return inVal
	}
	b := []byte(str)
	rune, _ := utf8.DecodeRune(b)
	rune = unicode.ToUpper(rune)
	// This assumes that the upper case rune is the same width as the lower case rune.
	// It's almost always true (might even be always).
	utf8.EncodeRune(b, rune)
	return stringValue(b)
}

func centerFilter(in Expr, c *Context, arg Expr) Value {
	inVal := in.Eval(c)
	count := int(arg.Eval(c).Int())
	if count <= 0 {
		return inVal
	}
	str := inVal.String()
	runes := []int(str)
	l := len(runes)
	if l >= count {
		return inVal
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

func cutFilter(in Expr, c *Context, arg Expr) Value {
	str := in.Eval(c).String()
	ch := arg.Eval(c).String()
	return stringValue(strings.Replace(str, ch, "", -1))
}

func dateFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func defaultFilter(in Expr, c *Context, arg Expr) Value {
	inVal := in.Eval(c)
	if b := inVal.Bool(); b {
		return inVal
	}
	return arg.Eval(c)
}

func defaultIfNilFilter(in Expr, c *Context, arg Expr) Value {
	inVal := in.Eval(c)
	if _, ok := inVal.(nilValue); !ok {
		return inVal
	}
	return arg.Eval(c)
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortreversedFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func divisiblebyFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

// TODO: This isn't quite right; this filter should work anywhere in a filter chain so it probably needs to be treated specially
func escapeFilter(in Expr, c *Context, arg Expr) Value {
	str := in.Eval(c).String()
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return stringValue(str)
}

func escapejsFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func filesizeformatFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

// Works on slices or arrays
func firstFilter(in Expr, c *Context, arg Expr) Value {
	inVal := in.Eval(c)
	v := inVal.Reflect()
	switch v.Kind() {
	case reflect.String:
		str := inVal.String()
		_, w := utf8.DecodeRuneInString(str)
		return stringValue(str[:w])
	case reflect.Array, reflect.Slice:
		return refToVal(v.Index(0))
	}
	return inVal
}

func fixAmpersandsFilter(in Expr, c *Context, arg Expr) Value {
	str := in.Eval(c).String()
	return stringValue(strings.Replace(str, "&", "&amp;", -1))
}

func floatformatFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func forceEscapeFilter(in Expr, c *Context, arg Expr) Value {
	str := in.Eval(c).String()
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return stringValue(str)
}

func getDigitFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func iriencodeFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func joinFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

// Works on slices or arrays
func lastFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Work on strings also
	inVal := in.Eval(c)
	list := inVal.Reflect()
	if list.Kind() != reflect.Array && list.Kind() != reflect.Slice {
		// TODO: Is this right?
		return inVal
	}
	len := list.Len()
	return refToVal(list.Index(len - 1))
}

func lengthFilter(in Expr, c *Context, arg Expr) Value {
	inVal := in.Eval(c)
	v := inVal.Reflect()
	switch v.Kind() {
	case reflect.String:
		return intValue(v.Len())
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return intValue(v.Len())
	}
	return inVal
}

func lengthIsFilter(in Expr, c *Context, arg Expr) Value {
	l := lengthFilter(in, c, nil).Int()
	return boolValue(l == arg.Eval(c).Int())
}

func linebreaksFilter(in Expr, c *Context, arg Expr) Value {
	str := in.Eval(c).String()
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "\n\n", "</p>", -1)
	str = strings.Replace(str, "\n", "<br />", -1)
	return stringValue(str)
}

func linebreaksbrFilter(in Expr, c *Context, arg Expr) Value {
	str := in.Eval(c).String()
	// TODO: We can probably get better performance by implementing this ourselves
	return stringValue(strings.Replace(str, "\n", "<br />", -1))
}

func linenumbersFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func ljustFilter(in Expr, c *Context, arg Expr) Value {
	inVal := in.Eval(c)
	count := int(arg.Eval(c).Int())
	if count <= 0 {
		return inVal
	}
	str := inVal.String()
	runes := []int(str)
	if len(runes) >= count {
		return inVal
	}
	count -= len(runes)
	return stringValue(str + strings.Repeat(" ", count))
}

func lowerFilter(in Expr, c *Context, arg Expr) Value {
	str := in.Eval(c).String()
	return stringValue(strings.ToLower(str))
}

func makeListFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func phone2numericFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func pluralizeFilter(in Expr, c *Context, arg Expr) Value {
	var single string
	var plural string
	suffix := arg.Eval(c).String()
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

	if in.Eval(c).Uint() > 0 {
		return stringValue(plural)
	}
	return stringValue(single)
}

func pprintFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func randomFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func removeTagsFilter(in Expr, c *Context, arg Expr) Value {
	// TODO: Implement
	return in.Eval(c)
}

func rjustFilter(in Expr, c *Context, arg Expr) Value {
	inVal := in.Eval(c)
	count := int(arg.Eval(c).Int())
	if count <= 0 {
		return inVal
	}
	str := inVal.String()
	runes := []int(str)
	if len(runes) >= count {
		return inVal
	}
	count -= len(runes)
	return stringValue(strings.Repeat(" ", count) + str)
}
