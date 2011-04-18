package template

import (
	"io"
	"reflect"
	"strconv"
)

// value can represent a variety of types. Basic types are represented as themselves,
// while composite types are represented as reflect.Value.
type value interface{}

func valueAsBool(v value) bool {
	switch v := v.(type) {
	case bool:
		return v
	case float32:
		return v != 0
	case float64:
		return v != 0
	case complex64:
		return v != 0
	case complex128:
		return v != 0
	case int:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case uintptr:
		return v != 0
	case string:
		return v != ""
	case reflect.Value:
		switch v.Kind() {
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
			return v.Len() != 0
		case reflect.Ptr:
			return !v.IsNil()
		case reflect.Struct:
			// TODO: Can we do any better?
			return true
		}
	}
	return false
}

// Convert value to string. Arrays, slices, and maps give Python-style output for Django compatibility.
func valueAsString(v value) string {
	switch v := v.(type) {
	case bool:
		if v {
			return "True"
		}
		return "False"
	case float32:
		return strconv.Ftoa32(v, 'g', -1)
	case float64:
		return strconv.Ftoa64(v, 'g', -1)
	case complex64:
		return strconv.Ftoa32(real(v), 'g', -1) + "+" + strconv.Ftoa32(imag(v), 'g', -1)
	case complex128:
		return strconv.Ftoa64(real(v), 'g', -1) + "+" + strconv.Ftoa64(imag(v), 'g', -1)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.Itoa64(v)
	case uint:
		return strconv.Uitoa(v)
	case uint8:
		return strconv.Uitoa(uint(v))
	case uint16:
		return strconv.Uitoa(uint(v))
	case uint32:
		return strconv.Uitoa(uint(v))
	case uint64:
		return strconv.Uitoa64(v)
	case uintptr:
		return strconv.Uitoa64(uint64(v))
	case string:
		return v
	case reflect.Value:
		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			str := "["
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					str += ", "
				}
				str1 := quoteString(refToVal(v.Index(i)))
				str += str1
			}
			str += "]"
			return str

		case reflect.Map:
			keys := v.MapKeys()
			str := "{"
			for i, key := range keys {
				if i > 0 {
					str += ", "
				}
				v1 := refToVal(key)
				str1 := quoteString(v1)
				str += str1
				str += ": "
				v1 = refToVal(v.MapIndex(key))
				str1 = quoteString(v1)
				str += str1
			}
			str += "}"
			return str

		case reflect.Ptr:
			// just print whatever the pointer points to
			return valueAsString(v.Elem())
		}
	}
	return ""
}

// If v is a string, put it in single quotes.
// Otherwise return the string normally.
func quoteString(v value) string {
	if s, ok := v.(string); ok {
		return "'" + s + "'"
	}
	return valueAsString(v)
}

func valueAsInt(v value) int64 {
	switch v := v.(type) {
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case uintptr:
		return int64(v)
	case string:
		u, err := strconv.Atoi64(v)
		if err == nil {
			return u
		}
	}
	return 0
}

func valueAsUint(v value) uint64 {
	switch v := v.(type) {
	case float32:
		return uint64(v)
	case float64:
		return uint64(v)
	case int:
		return uint64(v)
	case int8:
		return uint64(v)
	case int16:
		return uint64(v)
	case int32:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case uintptr:
		return uint64(v)
	case string:
		if u, err := strconv.Atoui64(v); err == nil {
			return u
		}
	}
	return 0
}

type valuer interface {
	Node
	value(s Stack) value
}

func renderValuer(v valuer, wr io.Writer, s Stack) {
	val := v.value(s)
	str := valueAsString(val)
	wr.Write([]byte(str))
}

type stringLit string

func (str stringLit) value(s Stack) value { return string(str) }

func (str stringLit) Render(wr io.Writer, s Stack) { wr.Write([]byte(string(str))) }

type intLit int64

func (i intLit) value(s Stack) value { return int64(i) }

func (i intLit) Render(wr io.Writer, s Stack) {
	str := strconv.Itoa64(int64(i))
	wr.Write([]byte(str))
}

type floatLit float64

func (f floatLit) value(s Stack) value { return float64(f) }

func (f floatLit) Render(wr io.Writer, s Stack) {
	str := strconv.Ftoa64(float64(f), 'g', -1)
	wr.Write([]byte(str))
}

// A variable represents a variable. Its integer is the index in the stack
// where it can be found.
type variable int

func (v variable) value(s Stack) value {
	return s[v]
}

func (v variable) set(val value, s Stack) {
	s[v] = val
}

func (v variable) Render(wr io.Writer, s Stack) {
	renderValuer(v, wr, s)
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
