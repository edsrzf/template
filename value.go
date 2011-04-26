package template

import (
	"io"
	"reflect"
	"strconv"
)

// If v is a string, put it in single quotes.
// Otherwise return the string normally.
func quoteString(v Value) string {
	if s, ok := v.(stringValue); ok {
		return "'" + string(s) + "'"
	}
	return v.String()
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
	Bool() bool

	// Int coerces the Value to a signed integer. The return parameter follows these
	// rules:
	//	- A bool that is true evaluates to 1; false evaluates to 0
	//	- An integer evaluates to itself, with unsigned types possibly overflowing
	//	- A string is converted to a signed integer if possible
	//	- A non-nil pointer to one of the above types uses 
	// All other values evaluate to 0.
	Int() int64

	// String coerces the Value to a string. The return parameter follows these rules:
	//	- A bool returns "true" if it is true and "false" otherwise
	//	- An integer is converted to its string representation
	//	- A string evaluates to itself
	//	- A non-nil pointer evaluates to whatever its element would evaluate to
	//	according to these rules
	// All other values evaluate to the empty string "".
	String() string

	// Uint coerces the Value to an unsigned integer. It uses the same rules as
	// Int except that negative signed integers will underflow to positive integers.
	Uint() uint64

	// Reflect returns the Value's reflected value.
	Reflect() reflect.Value
}

type nilValue byte

func (n nilValue) Bool() bool                      { return false }
func (n nilValue) Int() int64                      { return 0 }
func (n nilValue) String() string                  { return "" }
func (n nilValue) Uint() uint64                    { return 0 }
func (n nilValue) Reflect() reflect.Value          { return reflect.ValueOf(nil) }
func (n nilValue) Render(wr io.Writer, c *Context) {}

type boolValue bool

func (b boolValue) Bool() bool { return bool(b) }

func (b boolValue) Int() int64 {
	if b {
		return 1
	}
	return 0
}

func (b boolValue) String() string {
	if b {
		return "true"
	}
	return "false"
}

func (b boolValue) Uint() uint64 {
	if b {
		return 1
	}
	return 0
}

func (b boolValue) Reflect() reflect.Value          { return reflect.ValueOf(b) }
func (b boolValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(b.String())) }

type stringValue string

func (str stringValue) Bool() bool { return str != "" }

func (str stringValue) Int() int64 {
	if i, err := strconv.Atoi64(string(str)); err == nil {
		return i
	}
	return 0
}

func (str stringValue) String() string { return string(str) }

func (str stringValue) Uint() uint64 {
	if i, err := strconv.Atoui64(string(str)); err == nil {
		return i
	}
	return 0
}

func (str stringValue) Reflect() reflect.Value          { return reflect.ValueOf(str) }
func (str stringValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(string(str))) }

type intValue int64

func (i intValue) Bool() bool                      { return i != 0 }
func (i intValue) Int() int64                      { return int64(i) }
func (i intValue) String() string                  { return strconv.Itoa64(int64(i)) }
func (i intValue) Uint() uint64                    { return uint64(i) }
func (i intValue) Reflect() reflect.Value          { return reflect.ValueOf(i) }
func (i intValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(i.String())) }

type floatValue float64

func (f floatValue) Bool() bool             { return f != 0 }
func (f floatValue) Int() int64             { return int64(f) }
func (f floatValue) String() string         { return strconv.Ftoa64(float64(f), 'g', -1) }
func (f floatValue) Uint() uint64           { return uint64(f) }
func (f floatValue) Reflect() reflect.Value { return reflect.ValueOf(f) }

func (f floatValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(f.String())) }

type complexValue complex128

func (c complexValue) Bool() bool { return c != 0 }
func (c complexValue) Int() int64 { return 0 }

func (c complexValue) String() string {
	a, b := real(c), imag(c)
	return strconv.Ftoa64(a, 'g', -1) + "+" + strconv.Ftoa64(b, 'g', -1) + "i"
}

func (c complexValue) Uint() uint64           { return 0 }
func (c complexValue) Reflect() reflect.Value { return reflect.ValueOf(c) }

func (c complexValue) Render(wr io.Writer, _ *Context) {
	wr.Write([]byte(c.String()))
}

// reflectValue implements the common Value methods for reflected types.
type reflectValue reflect.Value

func (v reflectValue) Bool() bool             { return reflect.Value(v).Len() != 0 }
func (v reflectValue) Int() int64             { return 0 }
func (v reflectValue) Uint() uint64           { return 0 }
func (v reflectValue) Reflect() reflect.Value { return reflect.Value(v) }

// arrayValue represents a slice or array value
type arrayValue struct {
	reflectValue
}

func (a arrayValue) String() string {
	v := reflect.Value(a.reflectValue)
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
}
func (a arrayValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(a.String())) }

type mapValue struct {
	reflectValue
}

func (m mapValue) String() string {
	v := reflect.Value(m.reflectValue)
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
}
func (m mapValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(m.String())) }

type chanValue struct {
	reflectValue
}

func (ch chanValue) String() string {
	// TODO: implement
	return ""
}
func (ch chanValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(ch.String())) }

type structValue struct {
	reflectValue
}

func (st structValue) Bool() bool { return true }
func (st structValue) String() string {
	// TODO: implement
	return ""
}
func (st structValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(st.String())) }

type pointerValue struct {
	reflectValue
}

func (p pointerValue) value() Value { return refToVal(reflect.Value(p.reflectValue).Elem()) }
// TODO: correct
func (p pointerValue) Bool() bool { return !reflect.Value(p.reflectValue).IsNil() }
func (p pointerValue) String() string {
	if reflect.Value(p.reflectValue).IsNil() {
		return "<nil>"
	}
	return p.value().String()
}
func (p pointerValue) Render(wr io.Writer, c *Context) { wr.Write([]byte(p.String())) }

// A Variable is an index into a Context's stack.
// Variables must be obtained through the Parser before runtime.
type Variable int

func (v Variable) Eval(c *Context) Value {
	if val := c.stack[v]; val != nil {
		return val
	}
	return nilValue(0)
}

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
	case reflect.Complex64, reflect.Complex128:
		return complexValue(ref.Complex())
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

func lookup(v reflect.Value, s string) reflect.Value {
	var ret reflect.Value
	v = reflect.Indirect(v)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		if idx, err := strconv.Atoi(s); err == nil {
			ret = v.Index(idx)
		}
	case reflect.Map:
		keyt := v.Type().Key()
		switch keyt.Kind() {
		case reflect.String:
			ret = v.MapIndex(reflect.ValueOf(s))
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
	case reflect.Struct:
		ret = v.FieldByName(s)
	}
	// TODO: Find a way to look up methods by name
	return ret
}
