package template

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"utf8"
)

type filterFunc func(in value, context Context, arg valuer) value

type filter struct {
	f    filterFunc
	args valuer
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
	"add":             &regFilter{addFilter, ReqArg},
	"addslashes":      &regFilter{addslashesFilter, NoArg},
	"capfirst":        &regFilter{capfirstFilter, NoArg},
	"center":          &regFilter{centerFilter, ReqArg},
	"cut":             &regFilter{cutFilter, ReqArg},
	"default":         &regFilter{defaultFilter, ReqArg},
	"default_if_none": &regFilter{defaultIfNoneFilter, ReqArg},
	"escape":          &regFilter{escapeFilter, NoArg},
	"first":           &regFilter{firstFilter, NoArg},
	"lower":           &regFilter{lowerFilter, NoArg},
}

func addFilter(in value, context Context, arg valuer) value {
	l, ok1 := valueAsInt(in)
	r, ok2 := valueAsInt(arg.value(context))
	if !ok1 || !ok2 {
		return in
	}
	return l + r
}

func addslashesFilter(in value, context Context, arg valuer) value {
	str, _ := valueAsString(in)
	return strings.Replace(str, "'", "\\'", -1)
}

func capfirstFilter(in value, context Context, arg valuer) value {
	str, _ := valueAsString(in)
	if len(str) == 0 {
		return in
	}
	b := []byte(str)
	rune, _ := utf8.DecodeRune(b)
	rune = unicode.ToUpper(rune)
	// This assumes that the upper case rune is the same width as the lower case rune.
	// It's almost always true (might even be always).
	utf8.EncodeRune(b, rune)
	return string(b)
}

func centerFilter(in value, context Context, arg valuer) value {
	c, ok := valueAsInt(arg.value(context))
	count := int(c)
	if !ok || count <= 0 {
		return in
	}
	str, _ := valueAsString(in)
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
		return spaces + str + spaces
	}
	return strings.Repeat(" ", half) + str + strings.Repeat(" ", count)
}

// Our cutFilter is slightly more forgiving than Django's. It allows the argument to be an integer.
// {{ 123|cut:2 }} will output "13".
func cutFilter(in value, context Context, arg valuer) value {
	str, _ := valueAsString(in)
	ch, _ := valueAsString(arg.value(context))
	return strings.Replace(str, ch, "", -1)
}

func dateFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func defaultFilter(in value, context Context, arg valuer) value {
	b, _ := valueAsBool(in)
	if b {
		return in
	}
	def, _ := valueAsString(arg.value(context))
	return def
}

func defaultIfNoneFilter(in value, context Context, arg valuer) value {
	if in != nil {
		return in
	}
	def, _ := valueAsString(arg.value(context))
	return def
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortreversedFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func divisiblebyFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

// TODO: This isn't quite right; this filter should work anywhere in a filter chain so it probably needs to be treated specially
func escapeFilter(in value, context Context, arg valuer) value {
	str, _ := valueAsString(in)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return str
}

func escapejsFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func filesizeformatFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

// Works on slices or arrays
func firstFilter(in value, context Context, arg valuer) value {
	switch v := in.(type) {
	case string:
		_, w := utf8.DecodeRuneInString(v)
		return v[:w]
	case reflect.Value:
		k := v.Kind()
		if k == reflect.Array || k == reflect.Slice {
			return refToVal(v.Index(0))
		}
	}
	return in
}

func fixAmpersandsFilter(in value, context Context, arg valuer) value {
	str, _ := valueAsString(in)
	return strings.Replace(str, "&", "&amp;", -1)
}

func floatformatFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func forceEscapeFilter(in value, context Context, arg valuer) value {
	str, _ := valueAsString(in)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return str
}

func getDigitFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func iriencodeFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func joinFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

// Works on slices or arrays
func lastFilter(in value, context Context, arg valuer) value {
	// TODO: Work on strings also
	list, ok := in.(reflect.Value)
	if !ok || list.Kind() != reflect.Array && list.Kind() != reflect.Slice {
		// TODO: Is this right?
		return in
	}
	len := list.Len()
	return refToVal(list.Index(len - 1))
}

func lengthFilter(in value, context Context, arg valuer) value {
	switch v := in.(type) {
	case string:
		return len(v)
	case reflect.Value:
		k := v.Kind()
		switch k {
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
			return v.Len()
		}
	}
	return nil
}

func lengthIsFilter(in value, context Context, arg valuer) value {
	l, ok := lengthFilter(in, context, arg).(int)
	if !ok {
		// TODO: Is this right?
		return in
	}
	val, ok := valueAsInt(arg)
	if !ok {
		// TODO: Is this right?
		return in
	}
	return l == int(val)
}

func linebreaksFilter(in value, context Context, arg valuer) value {
	str, ok := in.(string)
	if !ok {
		return in
	}
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "\n\n", "</p>", -1)
	str = strings.Replace(str, "\n", "<br />", -1)
	return str
}

func linebreaksbrFilter(in value, context Context, arg valuer) value {
	str, ok := in.(string)
	if !ok {
		return in
	}
	// TODO: We can probably get better performance by implementing this ourselves
	return strings.Replace(str, "\n", "<br />", -1)
}

func linenumbersFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func ljustFilter(in value, context Context, arg valuer) value {
	c, ok := valueAsInt(arg.value(context))
	count := int(c)
	if !ok || count <= 0 {
		// TODO: Is this correct?
		return in
	}
	str := fmt.Sprint(in)
	runes := []int(str)
	if len(runes) >= count {
		// TODO: Is this correct?
		return in
	}
	count -= len(runes)
	return str + strings.Repeat(" ", count)
}

func lowerFilter(in value, context Context, arg valuer) value {
	str, ok := in.(string)
	if !ok {
		return in
	}
	str = strings.ToLower(str)
	return str
}

func makeListFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func phone2numericFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func pluralizeFilter(in value, context Context, arg valuer) value {
	var single string
	var plural string
	suffix, _ := valueAsString(arg.value(context))
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

	v, ok := valueAsUint(in)
	if !ok {
		// TODO: Is this right?
		return in
	}
	if v > 0 {
		return plural
	}
	return single
}

func pprintFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func randomFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func removeTagsFilter(in value, context Context, arg valuer) value {
	// TODO: Implement
	return in
}

func rjustFilter(in value, context Context, arg valuer) value {
	c, ok := valueAsInt(arg.value(context))
	count := int(c)
	if !ok || count <= 0 {
		// TODO: Is this correct?
		return in
	}
	str := fmt.Sprint(in)
	runes := []int(str)
	if len(runes) >= count {
		// TODO: Is this correct?
		return in
	}
	count -= len(runes)
	return strings.Repeat(" ", count) + str
}
