package template

import (
	"io"
	"reflect"
	"strconv"
)

// If v is a string, put it in single quotes.
// Otherwise return the string normally.
func quoteString(v Value, s Stack) string {
	if s, ok := v.(stringValue); ok {
		return "'" + string(s) + "'"
	}
	return v.String(s)
}

// TODO: should there be a Float method on Value?

// A Value represents a generic value of any type.
type Value interface {
	Node
	// Bool coerces the Value to a boolean. Values that are true are:
	//	- A bool that is true
	//	- A non-zero integer
	//	- A non-empty string
	//	- A slice, array, map, or channel with non-zero length
	//	- Any struct
	//	- A non-nil pointer to any of the above
	// All other values evaluate to false.
	Bool(s Stack) bool
	// Int coerces the Value to a signed integer. The return parameter follows these
	// rules:
	//	- A bool that is true evaluates to 1; false evaluates to 0
	//	- An integer evaluates to itself, with unsigned types possibly overflowing
	//	- A string is converted to a signed integer if possible
	//	- A non-nil pointer to one of the above types uses 
	// All other values evaluate to 0.
	Int(s Stack) int64
	// String coerces the Value to a string. The return parameter follows these rules:
	//	- A bool returns "true" if it is true and "false" otherwise
	//	- An integer is converted to its string representation
	//	- A string evaluates to itself
	//	- A non-nil pointer evaluates to whatever its element would evaluate to
	//	according to these rules
	// All other values evaluate to the empty string "".
	String(s Stack) string
	// Uint coerces the Value to an unsigned integer. It uses the same rules as
	// Int except that negative signed integers will underflow to positive integers.
	Uint(s Stack) uint64
	// Reflect returns the Value's reflected value.
	Reflect(s Stack) reflect.Value
}

type nilValue byte

func (n nilValue) Bool(s Stack) bool             { return false }
func (n nilValue) Int(s Stack) int64             { return 0 }
func (n nilValue) String(s Stack) string         { return "" }
func (n nilValue) Uint(s Stack) uint64           { return 0 }
func (n nilValue) Reflect(s Stack) reflect.Value { return reflect.NewValue(nil) }
func (n nilValue) Render(wr io.Writer, s Stack)  {}

type boolValue bool

func (b boolValue) Bool(s Stack) bool { return bool(b) }
func (b boolValue) Int(s Stack) int64 {
	if b {
		return 1
	}
	return 0
}
func (b boolValue) String(s Stack) string {
	if b {
		return "true"
	}
	return "false"
}
func (b boolValue) Uint(s Stack) uint64 {
	if b {
		return 1
	}
	return 0
}
func (b boolValue) Reflect(s Stack) reflect.Value { return reflect.NewValue(b) }
func (b boolValue) Render(wr io.Writer, s Stack)  { wr.Write([]byte(b.String(s))) }

// Generic function that works for any Value. Some specific Values can do this faster.
func renderValue(v Value, wr io.Writer, s Stack) {
	wr.Write([]byte(v.String(s)))
}

type stringValue string

func (str stringValue) Bool(s Stack) bool { return str != "" }

func (str stringValue) Int(s Stack) int64 {
	if i, err := strconv.Atoi64(string(str)); err == nil {
		return i
	}
	return 0
}

func (str stringValue) String(s Stack) string { return string(str) }

func (str stringValue) Uint(s Stack) uint64 {
	if i, err := strconv.Atoui64(string(str)); err == nil {
		return i
	}
	return 0
}

func (str stringValue) Reflect(s Stack) reflect.Value { return reflect.NewValue(str) }

func (str stringValue) Render(wr io.Writer, s Stack) { wr.Write([]byte(string(str))) }

type intValue int64

func (i intValue) Bool(s Stack) bool             { return i != 0 }
func (i intValue) Int(s Stack) int64             { return int64(i) }
func (i intValue) String(s Stack) string         { return strconv.Itoa64(int64(i)) }
func (i intValue) Uint(s Stack) uint64           { return uint64(i) }
func (i intValue) Reflect(s Stack) reflect.Value { return reflect.NewValue(i) }

func (i intValue) Render(wr io.Writer, s Stack) {
	wr.Write([]byte(i.String(s)))
}

type floatValue float64

func (f floatValue) Bool(s Stack) bool             { return f != 0 }
func (f floatValue) Int(s Stack) int64             { return int64(f) }
func (f floatValue) String(s Stack) string         { return strconv.Ftoa64(float64(f), 'g', -1) }
func (f floatValue) Uint(s Stack) uint64           { return uint64(f) }
func (f floatValue) Reflect(s Stack) reflect.Value { return reflect.NewValue(f) }

func (f floatValue) Render(wr io.Writer, s Stack) {
	wr.Write([]byte(f.String(s)))
}

/*
TODO: uncomment this when issue 1716 is fixed
type complexValue complex128

func (c complexValue) Bool(s Stack) bool { return c != 0 }
func (c complexValue) Int(s Stack) bool { return 0 }
// TODO: implement
func (c complexValue) String(s Stack) bool { return "" }
func (c complexValue) Uint(s Stack) bool { return 0 }
func (c complexValue) Reflect(s Stack) reflect.Value { return reflect.NewValue(c) }

func (c complexValue) Render(wr io.Writer, s Stack) {
	wr.Write([]byte(c.String(s)))
}
*/

// reflectValue implements the common Value methods for reflected types.
type reflectValue reflect.Value

func (v reflectValue) Bool(s Stack) bool             { return reflect.Value(v).Len() != 0 }
func (v reflectValue) Int(s Stack) int64             { return 0 }
func (v reflectValue) Uint(s Stack) uint64           { return 0 }
func (v reflectValue) Reflect(s Stack) reflect.Value { return reflect.Value(v) }

// arrayValue represents a slice or array value
type arrayValue struct {
	reflectValue
}

func (a arrayValue) String(s Stack) string {
	v := reflect.Value(a.reflectValue)
	str := "["
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			str += ", "
		}
		str1 := quoteString(refToVal(v.Index(i)), s)
		str += str1
	}
	str += "]"
	return str
}
func (a arrayValue) Render(wr io.Writer, s Stack) { renderValue(a, wr, s) }

type mapValue struct {
	reflectValue
}

func (m mapValue) String(s Stack) string {
	v := reflect.Value(m.reflectValue)
	keys := v.MapKeys()
	str := "{"
	for i, key := range keys {
		if i > 0 {
			str += ", "
		}
		v1 := refToVal(key)
		str1 := quoteString(v1, s)
		str += str1
		str += ": "
		v1 = refToVal(v.MapIndex(key))
		str1 = quoteString(v1, s)
		str += str1
	}
	str += "}"
	return str
}
func (m mapValue) Render(wr io.Writer, s Stack) { renderValue(m, wr, s) }

type chanValue struct {
	reflectValue
}

func (c chanValue) String(s Stack) string {
	// TODO: implement
	return ""
}
func (c chanValue) Render(wr io.Writer, s Stack) { renderValue(c, wr, s) }

type structValue struct {
	reflectValue
}

func (st structValue) Bool(s Stack) bool { return true }
func (st structValue) String(s Stack) string {
	// TODO: implement
	return ""
}
func (st structValue) Render(wr io.Writer, s Stack) { renderValue(st, wr, s) }

type pointerValue struct {
	reflectValue
}

func (p pointerValue) value() Value { return refToVal(reflect.Value(p.reflectValue).Elem()) }
// TODO: correct
func (p pointerValue) Bool(s Stack) bool { return !reflect.Value(p.reflectValue).IsNil() }
func (p pointerValue) String(s Stack) string {
	if reflect.Value(p.reflectValue).IsNil() {
		return "<nil>"
	}
	return p.value().String(s)
}
func (p pointerValue) Render(wr io.Writer, s Stack) { renderValue(p, wr, s) }

// A Variable is an index into a Template's runtime Stack.
type Variable int

func (v Variable) Bool(s Stack) bool {
	if val := s[v]; val != nil {
		return val.Bool(s)
	}
	return false
}

func (v Variable) Int(s Stack) int64 {
	if val := s[v]; val != nil {
		return val.Int(s)
	}
	return 0
}

func (v Variable) String(s Stack) string {
	if val := s[v]; val != nil {
		return val.String(s)
	}
	return ""
}

func (v Variable) Uint(s Stack) uint64 {
	if val := s[v]; val != nil {
		return val.Uint(s)
	}
	return 0
}

func (v Variable) Reflect(s Stack) reflect.Value {
	if val := s[v]; val != nil {
		return val.Reflect(s)
	}
	return reflect.NewValue(nil)
}

func (v Variable) Render(wr io.Writer, s Stack) { renderValue(v, wr, s) }

func (v Variable) Set(val Value, s Stack) { s[v] = val }

func getVal(ref reflect.Value, specs []string) Value {
	for _, s := range specs {
		ref = lookup(ref, s)
		if ref.Kind() == reflect.Invalid {
			return nilValue(0)
		}
	}
	return refToVal(ref)
}

func refToVal(ref reflect.Value) Value {
	switch ref.Kind() {
	case reflect.Bool:
		return boolValue(ref.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intValue(ref.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		return intValue(ref.Uint())
	case reflect.Float32, reflect.Float64:
		return floatValue(ref.Float())
	// TODO: uncomment this when issue 1716 is fixed
	//case reflect.Complex64, reflect.Complex128:
	//return complexValue(ref.Complex())
	case reflect.Array, reflect.Slice:
		return arrayValue{reflectValue(ref)}
	case reflect.Chan:
		return chanValue{reflectValue(ref)}
	case reflect.Map:
		return mapValue{reflectValue(ref)}
	case reflect.Ptr:
		return pointerValue{reflectValue(ref)}
	case reflect.String:
		return stringValue(ref.String())
	case reflect.Struct:
		return structValue{reflectValue(ref)}
	}
	return nilValue(0)
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
				idxVal := reflect.New(keyt).Elem()
				idxVal.SetInt(idx)
				ret = v.MapIndex(idxVal)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if idx, err := strconv.Atoui64(s); err == nil {
				idxVal := reflect.New(keyt).Elem()
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
