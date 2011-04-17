/*type if {
   cond condition
   ifNode node
   elseNode node
}

func (i *if) render(wr io.Writer, context map[string]interface{}) {
   if condition.eval(context) {
        ifNode.render(wr, context)
   } else {
        elseNode.render(wr, context)
   }
}*/
package template

import (
	"reflect"
	"strconv"
	"utf8"
)

type value interface{}

func valueAsBool(v value) (bool, bool) {
	switch v := v.(type) {
	case bool:
		return v, true
	case float32:
		return v != 0, true
	case float64:
		return v != 0, true
	case complex64:
		return v != 0, true
	case complex128:
		return v != 0, true
	case int:
		return v != 0, true
	case int8:
		return v != 0, true
	case int16:
		return v != 0, true
	case int32:
		return v != 0, true
	case int64:
		return v != 0, true
	case uint:
		return v != 0, true
	case uint8:
		return v != 0, true
	case uint16:
		return v != 0, true
	case uint32:
		return v != 0, true
	case uint64:
		return v != 0, true
	case uintptr:
		return v != 0, true
	case string:
		return v != "", true
	case reflect.Value:
		switch v.Kind() {
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
			return v.Len() != 0, true
		case reflect.Ptr:
			return !v.IsNil(), true
		case reflect.Struct:
			// TODO: Can we do any better?
			return true, true
		}
	}
	return false, false
}

// Convert value to string. Arrays, slices, and maps give Python-style output for Django compatibility.
func valueAsString(v value) (string, bool) {
	switch v := v.(type) {
	case bool:
		if v {
			return "True", true
		} else {
			return "False", true
		}
	case float32:
		return strconv.Ftoa32(v, 'g', -1), true
	case float64:
		return strconv.Ftoa64(v, 'g', -1), true
	case complex64:
		return strconv.Ftoa32(real(v), 'g', -1) + "+" + strconv.Ftoa32(imag(v), 'g', -1), true
	case complex128:
		return strconv.Ftoa64(real(v), 'g', -1) + "+" + strconv.Ftoa64(imag(v), 'g', -1), true
	case int:
		return strconv.Itoa(v), true
	case int8:
		return strconv.Itoa(int(v)), true
	case int16:
		return strconv.Itoa(int(v)), true
	case int32:
		return strconv.Itoa(int(v)), true
	case int64:
		return strconv.Itoa64(v), true
	case uint:
		return strconv.Uitoa(v), true
	case uint8:
		return strconv.Uitoa(uint(v)), true
	case uint16:
		return strconv.Uitoa(uint(v)), true
	case uint32:
		return strconv.Uitoa(uint(v)), true
	case uint64:
		return strconv.Uitoa64(v), true
	case uintptr:
		return strconv.Uitoa64(uint64(v)), true
	case string:
		return v, true
	case reflect.Value:
		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			str := "["
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					str += ", "
				}
				str1, ok := quoteString(refToVal(v.Index(i)))
				if !ok {
					return "", false
				}
				str += str1
			}
			str += "]"
			return str, true

		case reflect.Map:
			keys := v.MapKeys()
			str := "{"
			for i, key := range keys {
				if i > 0 {
					str += ", "
				}
				v1 := refToVal(key)
				str1, ok := quoteString(v1)
				if !ok {
					return "", false
				}
				str += str1
				str += ": "
				v1 = refToVal(v.MapIndex(key))
				str1, ok = quoteString(v1)
				if !ok {
					return "", false
				}
				str += str1
			}
			str += "}"
			return str, true

		case reflect.Ptr:
			// just print whatever the pointer points to
			return valueAsString(v.Elem())
		}
	}
	return "", false
}

// If v is a string, put it in single quotes.
// Otherwise return the string normally.
func quoteString(v value) (string, bool) {
	if s, ok := v.(string); ok {
		return "'" + s + "'", true
	}
	return valueAsString(v)
}

func valueAsInt(v value) (int64, bool) {
	switch v := v.(type) {
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case uintptr:
		return int64(v), true
	case string:
		u, err := strconv.Atoi64(v)
		if err == nil {
			return u, true
		}
	}
	return 0, false
}

func valueAsUint(v value) (uint64, bool) {
	switch v := v.(type) {
	case float32:
		return uint64(v), true
	case float64:
		return uint64(v), true
	case int:
		return uint64(v), true
	case int8:
		return uint64(v), true
	case int16:
		return uint64(v), true
	case int32:
		return uint64(v), true
	case int64:
		return uint64(v), true
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	case uintptr:
		return uint64(v), true
	case string:
		if u, err := strconv.Atoui64(v); err == nil {
			return u, true
		}
	}
	return 0, false
}

type condition interface {
	eval(context Context) bool
}

type equal struct {
	left, right *variable
}

func (e *equal) eval(context Context) bool {
	// TODO: Make sure types are comparable
	l := e.left.value(context)
	r := e.right.value(context)
	return l == r
}

type nequal struct {
	left, right *variable
}

func (n *nequal) eval(context Context) bool {
	// TODO: Make sure types are comparable
	l := n.left.value(context)
	r := n.right.value(context)
	return l != r
}

type not struct {
	inner condition
}

func (n *not) eval(context Context) bool {
	return !n.inner.eval(context)
}

type and struct {
	left, right condition
}

func (a *and) eval(context Context) bool {
	return a.left.eval(context) && a.right.eval(context)
}

type or struct {
	left, right condition
}

func (o *or) eval(context Context) bool {
	return o.left.eval(context) || o.right.eval(context)
}

type valuer interface {
	value(context Context) value
}

type stringLit string

func (s stringLit) value(context Context) value {
	return string(s)
}

type intLit int64

func (i intLit) value(context Context) value {
	return int64(i)
}

type floatLit float64

func (f floatLit) value(context Context) value {
	return float64(f)
}

type variable []string

func (v variable) value(context Context) value {
	// TODO: Panic instead?
	if len(v) == 0 {
		return nil
	}
	var val value
	if context != nil {
		val = context[v[0]]
	}
	if val == nil {
		return nil
	}
	// We might be able to avoid reflection
	var ret value
	switch val := val.(type) {
	case bool:
		ret = val
	case float32:
		ret = val
	case float64:
		ret = val
	case complex64:
		ret = val
	case complex128:
		ret = val
	case int:
		ret = val
	case int8:
		ret = val
	case int16:
		ret = val
	case int32:
		ret = val
	case int64:
		ret = val
	case uint:
		ret = val
	case uint8:
		ret = val
	case uint16:
		ret = val
	case uint32:
		ret = val
	case uint64:
		ret = val
	case uintptr:
		ret = val
	case string:
		ret = val
	}
	if ret != nil {
		// This has an attribute specified, but only strings accept one.
		if len(v) > 1 {
			str, ok := ret.(string)
			if !ok {
				return nil
			}
			idx, err := strconv.Atoi(v[1])
			if err != nil {
				return nil
			}
			var n, i, c int
			for i, c = range str {
				if n == idx {
					break
				}
				n++
			}
			return str[i : i+utf8.RuneLen(c)]
		}
	} else {
		ref := reflect.NewValue(val)
		ret = getVal(ref, v[1:])
	}
	return ret
}

// Represents a variable with possible attributes accessed.
// Attributes work like this:
// For the expression "a.b"
// - If a is a map[string]T, this is treated as a["b"]
// - If a is a struct or pointer to a struct, this is treated as a.b
// - If a is a map[numeric]T, slice, array, or pointer to an array, this is treated as a[b]
// - If the above all fail, this is treated as a method call a.b()
type expr struct {
	// strings that are syntactically separated by a dot
	v       valuer
	filters []*filter
}

// Returns a true or false value for a variable
// false values include:
// - nil
// - A slice, map, channel, array, or pointer to an array with zero len
// - The bool value false
// - An empty string
// - Zero of any numeric type
func (e *expr) eval(context Context) bool {
	v := e.value(context)
	if v == nil {
		return false
	}

	ret, _ := valueAsBool(v)
	return ret
}

func (e *expr) value(context Context) value {
	ret := e.v.value(context)
	for _, f := range e.filters {
		ret = f.f(ret, context, f.args)
	}
	return ret
}

func getVal(ref reflect.Value, specs []string) value {
	for _, s := range specs {
		ref = lookup(ref, s)
		if ref.Kind() == reflect.Invalid {
			return nil
		}
	}
	return refToVal(ref)
}

func refToVal(ref reflect.Value) value {
	k := ref.Kind()
	if reflect.Bool <= k && k <= reflect.Complex128 || k == reflect.String {
		return ref.Interface()
	}
	return ref

}

func listElem(v reflect.Value, s string) reflect.Value {
	if idx, err := strconv.Atoi(s); err == nil {
		return v.Index(idx)
	}
	return reflect.Value{}
}

func lookup(v reflect.Value, s string) reflect.Value {
	var ret reflect.Value
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		ret = listElem(v, s)
	case reflect.Map:
		keyt := v.Type().Key()
		switch keyt.Kind() {
		case reflect.String:
			ret = v.MapIndex(reflect.NewValue(s))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if idx, err := strconv.Atoi64(s); err == nil {
				idxVal := reflect.Zero(keyt)
				idxVal.SetInt(idx)
				ret = v.MapIndex(idxVal)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if idx, err := strconv.Atoui64(s); err == nil {
				idxVal := reflect.Zero(keyt)
				idxVal.SetUint(idx)
				ret = v.MapIndex(idxVal)
			}
		}
	case reflect.Ptr:
		v := v.Elem()
		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			ret = listElem(v, s)
		case reflect.Struct:
			ret = v.FieldByName(s)
		}
	case reflect.Struct:
		ret = v.FieldByName(s)
	}
	// TODO: Find a way to look up methods by name
	return ret
}
