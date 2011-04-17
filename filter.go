package template

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"utf8"
)

type filterFunc func(in value, s Stack, arg valuer) value

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
	"default_if_nil":  &regFilter{defaultIfNilFilter, ReqArg},
	"escape":          &regFilter{escapeFilter, NoArg},
	"first":           &regFilter{firstFilter, NoArg},
	"lower":           &regFilter{lowerFilter, NoArg},
}

func addFilter(in value, s Stack, arg valuer) value {
	l := valueAsInt(in)
	r := valueAsInt(arg.value(s))
	return l + r
}

func addslashesFilter(in value, s Stack, arg valuer) value {
	str := valueAsString(in)
	return strings.Replace(str, "'", "\\'", -1)
}

func capfirstFilter(in value, s Stack, arg valuer) value {
	str := valueAsString(in)
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

func centerFilter(in value, s Stack, arg valuer) value {
	c := valueAsInt(arg.value(s))
	count := int(c)
	if count <= 0 {
		return in
	}
	str := valueAsString(in)
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
func cutFilter(in value, s Stack, arg valuer) value {
	str := valueAsString(in)
	ch := valueAsString(arg.value(s))
	return strings.Replace(str, ch, "", -1)
}

func dateFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func defaultFilter(in value, s Stack, arg valuer) value {
	if b := valueAsBool(in); b {
		return in
	}
	return valueAsString(arg.value(s))
}

func defaultIfNilFilter(in value, s Stack, arg valuer) value {
	if in != nil {
		return in
	}
	return valueAsString(arg.value(s))
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

// Instead of taking a list of dictionaries, it takes a slice of maps
func dictsortreversedFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func divisiblebyFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

// TODO: This isn't quite right; this filter should work anywhere in a filter chain so it probably needs to be treated specially
func escapeFilter(in value, s Stack, arg valuer) value {
	str := valueAsString(in)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return str
}

func escapejsFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func filesizeformatFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

// Works on slices or arrays
func firstFilter(in value, s Stack, arg valuer) value {
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

func fixAmpersandsFilter(in value, s Stack, arg valuer) value {
	str := valueAsString(in)
	return strings.Replace(str, "&", "&amp;", -1)
}

func floatformatFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func forceEscapeFilter(in value, s Stack, arg valuer) value {
	str := valueAsString(in)
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "'", "&#39;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)
	return str
}

func getDigitFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func iriencodeFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func joinFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

// Works on slices or arrays
func lastFilter(in value, s Stack, arg valuer) value {
	// TODO: Work on strings also
	list, ok := in.(reflect.Value)
	if !ok || list.Kind() != reflect.Array && list.Kind() != reflect.Slice {
		// TODO: Is this right?
		return in
	}
	len := list.Len()
	return refToVal(list.Index(len - 1))
}

func lengthFilter(in value, s Stack, arg valuer) value {
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

func lengthIsFilter(in value, s Stack, arg valuer) value {
	l, ok := lengthFilter(in, s, arg).(int)
	if !ok {
		return false
	}
	return int64(l) == valueAsInt(arg)
}

func linebreaksFilter(in value, s Stack, arg valuer) value {
	str, ok := in.(string)
	if !ok {
		return in
	}
	// TODO: We can probably get better performance by implementing this ourselves
	str = strings.Replace(str, "\n\n", "</p>", -1)
	str = strings.Replace(str, "\n", "<br />", -1)
	return str
}

func linebreaksbrFilter(in value, s Stack, arg valuer) value {
	str, ok := in.(string)
	if !ok {
		return in
	}
	// TODO: We can probably get better performance by implementing this ourselves
	return strings.Replace(str, "\n", "<br />", -1)
}

func linenumbersFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func ljustFilter(in value, s Stack, arg valuer) value {
	count := int(valueAsInt(arg.value(s)))
	if count <= 0 {
		return in
	}
	str := fmt.Sprint(in)
	runes := []int(str)
	if len(runes) >= count {
		return in
	}
	count -= len(runes)
	return str + strings.Repeat(" ", count)
}

func lowerFilter(in value, s Stack, arg valuer) value {
	str, ok := in.(string)
	if !ok {
		return in
	}
	str = strings.ToLower(str)
	return str
}

func makeListFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func phone2numericFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func pluralizeFilter(in value, s Stack, arg valuer) value {
	var single string
	var plural string
	suffix := valueAsString(arg.value(s))
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

	if valueAsUint(in) > 0 {
		return plural
	}
	return single
}

func pprintFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func randomFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func removeTagsFilter(in value, s Stack, arg valuer) value {
	// TODO: Implement
	return in
}

func rjustFilter(in value, s Stack, arg valuer) value {
	count := int(valueAsInt(arg.value(s)))
	if count <= 0 {
		return in
	}
	str := fmt.Sprint(in)
	runes := []int(str)
	if len(runes) >= count {
		return in
	}
	count -= len(runes)
	return strings.Repeat(" ", count) + str
}
