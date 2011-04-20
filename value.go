package template

import (
	"io"
	"reflect"
	"strconv"
)

// If v is a string, put it in single quotes.
// Otherwise return the string normally.
func quoteString(v Value, c *Context) string {
	if s, ok := v.(stringValue); ok {
		return "'" + string(s) + "'"
	}
	return v.String(c)
}

// TODO: should there be a Float method on Value?

// A Value represents a generic value of any type.
type Value interface {
	Node
	// TODO: the Eval documentation really sucks

	// Eval evaluates the Value, performing any necessary operations to get the
	// final Value on which Value's other methods operate.
	// This method should be used to avoid unnecessary multiple evaluations of the Value.
	Eval(c *Context) Value
	// Bool coerces the Value to a boolean. Values that are true are:
	//	- A bool that is true
	//	- A non-zero integer
	//	- A non-empty string
	//	- A slice, array, map, or channel with non-zero length
	//	- Any struct
	//	- A non-nil pointer to any of the above
	// All other values evaluate to false.
	Bool(c *Context) bool
	// Int coerces the Value to a signed integer. The return parameter follows these
	// rules:
	//	- A bool that is true evaluates to 1; false evaluates to 0
	//	- An integer evaluates to itself, with unsigned types possibly overflowing
	//	- A string is converted to a signed integer if possible
	//	- A non-nil pointer to one of the above types uses 
	// All other values evaluate to 0.
	Int(c *Context) int64
	// String coerces the Value to a string. The return parameter follows these rules:
	//	- A bool returns "true" if it is true and "false" otherwise
	//	- An integer is converted to its string representation
	//	- A string evaluates to itself
	//	- A non-nil pointer evaluates to whatever its element would evaluate to
	//	according to these rules
	// All other values evaluate to the empty string "".
	String(c *Context) string
	// Uint coerces the Value to an unsigned integer. It uses the same rules as
	// Int except that negative signed integers will underflow to positive integers.
	Uint(c *Context) uint64
	// Reflect returns the Value's reflected value.
	Reflect(c *Context) reflect.Value
}

type nilValue byte

func (n nilValue) Eval(c *Context) Value            { return n }
func (n nilValue) Bool(c *Context) bool             { return false }
func (n nilValue) Int(c *Context) int64             { return 0 }
func (n nilValue) String(c *Context) string         { return "" }
func (n nilValue) Uint(c *Context) uint64           { return 0 }
func (n nilValue) Reflect(c *Context) reflect.Value { return reflect.NewValue(nil) }
func (n nilValue) Render(wr io.Writer, c *Context)  {}

type boolValue bool

func (b boolValue) Eval(c *Context) Value { return b }
func (b boolValue) Bool(c *Context) bool  { return bool(b) }
func (b boolValue) Int(c *Context) int64 {
	if b {
		return 1
	}
	return 0
}
func (b boolValue) String(c *Context) string {
	if b {
		return "true"
	}
	return "false"
}
func (b boolValue) Uint(c *Context) uint64 {
	if b {
		return 1
	}
	return 0
}
func (b boolValue) Reflect(c *Context) reflect.Value { return reflect.NewValue(b) }
func (b boolValue) Render(wr io.Writer, c *Context)  { wr.Write([]byte(b.String(c))) }

// Generic function that works for any Value. Some specific Values can do this faster.
func renderValue(v Value, wr io.Writer, c *Context) {
	wr.Write([]byte(v.String(c)))
}

type stringValue string

func (str stringValue) Eval(c *Context) Value { return str }
func (str stringValue) Bool(c *Context) bool  { return str != "" }

func (str stringValue) Int(c *Context) int64 {
	if i, err := strconv.Atoi64(string(str)); err == nil {
		return i
	}
	return 0
}

func (str stringValue) String(c *Context) string { return string(str) }

func (str stringValue) Uint(c *Context) uint64 {
	if i, err := strconv.Atoui64(string(str)); err == nil {
		return i
	}
	return 0
}

func (str stringValue) Reflect(c *Context) reflect.Value { return reflect.NewValue(str) }

func (str stringValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(string(str))) }

type intValue int64

func (i intValue) Bool(c *Context) bool             { return i != 0 }
func (i intValue) Int(c *Context) int64             { return int64(i) }
func (i intValue) String(c *Context) string         { return strconv.Itoa64(int64(i)) }
func (i intValue) Uint(c *Context) uint64           { return uint64(i) }
func (i intValue) Reflect(c *Context) reflect.Value { return reflect.NewValue(i) }
func (i intValue) Eval(c *Context) Value            { return i }

func (i intValue) Render(wr io.Writer, c *Context) {
	wr.Write([]byte(i.String(c)))
}

type floatValue float64

func (f floatValue) Bool(c *Context) bool             { return f != 0 }
func (f floatValue) Int(c *Context) int64             { return int64(f) }
func (f floatValue) String(c *Context) string         { return strconv.Ftoa64(float64(f), 'g', -1) }
func (f floatValue) Uint(c *Context) uint64           { return uint64(f) }
func (f floatValue) Reflect(c *Context) reflect.Value { return reflect.NewValue(f) }
func (f floatValue) Eval(c *Context) Value            { return f }

func (f floatValue) Render(wr io.Writer, c *Context) {
	wr.Write([]byte(f.String(c)))
}

/*
TODO: uncomment this when issue 1716 is fixed
type complexValue complex128

func (c complexValue) Bool(c *Context) bool { return c != 0 }
func (c complexValue) Int(c *Context) bool { return 0 }
// TODO: implement
func (c complexValue) String(c *Context) bool { return "" }
func (c complexValue) Uint(c *Context) bool { return 0 }
func (c complexValue) Reflect(c *Context) reflect.Value { return reflect.NewValue(c) }

func (c complexValue) Render(wr io.Writer, c *Context) {
	wr.Write([]byte(c.String(c)))
}
*/

// reflectValue implements the common Value methods for reflected types.
type reflectValue reflect.Value

func (v reflectValue) Bool(c *Context) bool             { return reflect.Value(v).Len() != 0 }
func (v reflectValue) Int(c *Context) int64             { return 0 }
func (v reflectValue) Uint(c *Context) uint64           { return 0 }
func (v reflectValue) Reflect(c *Context) reflect.Value { return reflect.Value(v) }

// arrayValue represents a slice or array value
type arrayValue struct {
	reflectValue
}

func (a arrayValue) String(c *Context) string {
	v := reflect.Value(a.reflectValue)
	str := "["
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			str += ", "
		}
		str1 := quoteString(refToVal(v.Index(i)), c)
		str += str1
	}
	str += "]"
	return str
}
func (a arrayValue) Eval(c *Context) Value           { return a }
func (a arrayValue) Render(wr io.Writer, c *Context) { renderValue(a, wr, c) }

type mapValue struct {
	reflectValue
}

func (m mapValue) String(c *Context) string {
	v := reflect.Value(m.reflectValue)
	keys := v.MapKeys()
	str := "{"
	for i, key := range keys {
		if i > 0 {
			str += ", "
		}
		v1 := refToVal(key)
		str1 := quoteString(v1, c)
		str += str1
		str += ": "
		v1 = refToVal(v.MapIndex(key))
		str1 = quoteString(v1, c)
		str += str1
	}
	str += "}"
	return str
}
func (m mapValue) Eval(c *Context) Value           { return m }
func (m mapValue) Render(wr io.Writer, c *Context) { renderValue(m, wr, c) }

type chanValue struct {
	reflectValue
}

func (ch chanValue) String(c *Context) string {
	// TODO: implement
	return ""
}
func (ch chanValue) Eval(c *Context) Value           { return ch }
func (ch chanValue) Render(wr io.Writer, c *Context) { renderValue(ch, wr, c) }

type structValue struct {
	reflectValue
}

func (st structValue) Bool(c *Context) bool { return true }
func (st structValue) String(c *Context) string {
	// TODO: implement
	return ""
}
func (st structValue) Eval(c *Context) Value           { return st }
func (st structValue) Render(wr io.Writer, c *Context) { renderValue(st, wr, c) }

type pointerValue struct {
	reflectValue
}

func (p pointerValue) value() Value { return refToVal(reflect.Value(p.reflectValue).Elem()) }
// TODO: correct
func (p pointerValue) Bool(c *Context) bool { return !reflect.Value(p.reflectValue).IsNil() }
func (p pointerValue) String(c *Context) string {
	if reflect.Value(p.reflectValue).IsNil() {
		return "<nil>"
	}
	return p.value().String(c)
}
func (p pointerValue) Eval(c *Context) Value           { return p }
func (p pointerValue) Render(wr io.Writer, c *Context) { renderValue(p, wr, c) }

// A Variable is an index into a Context's stack.
// Variables must be obtained through the Parser before runtime.
type Variable int

func (v Variable) Bool(c *Context) bool {
	if val := c.stack[v]; val != nil {
		return val.Bool(c)
	}
	return false
}

func (v Variable) Int(c *Context) int64 {
	if val := c.stack[v]; val != nil {
		return val.Int(c)
	}
	return 0
}

func (v Variable) String(c *Context) string {
	if val := c.stack[v]; val != nil {
		return val.String(c)
	}
	return ""
}

func (v Variable) Uint(c *Context) uint64 {
	if val := c.stack[v]; val != nil {
		return val.Uint(c)
	}
	return 0
}

func (v Variable) Reflect(c *Context) reflect.Value {
	if val := c.stack[v]; val != nil {
		return val.Reflect(c)
	}
	return reflect.NewValue(nil)
}

func (v Variable) Eval(c *Context) Value {
	if val := c.stack[v]; val != nil {
		return val
	}
	return nilValue(0)
}

func (v Variable) Render(wr io.Writer, c *Context) { renderValue(v, wr, c) }

func (v Variable) Set(val Value, c *Context) { c.stack[v] = val }

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
